package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Fetcher downloads a base project's .agents/ directory into a local destination.
// Implement this interface to support additional source schemes.
type Fetcher interface {
	// CanHandle reports whether this fetcher handles the given source string.
	CanHandle(source string) bool
	// FetchAgentsDir downloads the base source's .agents/ directory and writes
	// all files into destDir. destDir already exists when this is called.
	FetchAgentsDir(source, destDir string) error
}

var fetchers = []Fetcher{
	&localFetcher{},
	&httpFetcher{},
	&gitFetcher{},
	&ftpFetcher{},
}

// RegisterFetcher prepends f to the fetcher chain so it is tried first.
// Use this in tests or to add support for custom source schemes.
func RegisterFetcher(f Fetcher) {
	fetchers = append([]Fetcher{f}, fetchers...)
}

// CanFetchSource reports whether any registered fetcher can handle source.
func CanFetchSource(source string) bool {
	for _, f := range fetchers {
		if f.CanHandle(source) {
			return true
		}
	}
	return false
}

// FetchAgentsDir dispatches to the appropriate fetcher for source.
func FetchAgentsDir(source, destDir string) error {
	for _, f := range fetchers {
		if f.CanHandle(source) {
			return f.FetchAgentsDir(source, destDir)
		}
	}
	return fmt.Errorf("no fetcher for %q — supported schemes: local path, https://, git@.../https://….git, ftp://", source)
}

// collectFilePaths returns all file paths referenced in cfg across every field.
func collectFilePaths(cfg *Config) []string {
	var paths []string
	paths = append(paths, cfg.Rules...)
	paths = append(paths, cfg.Skills...)
	paths = append(paths, cfg.Context...)
	paths = append(paths, cfg.Commands...)
	for _, p := range cfg.Personas {
		if p.Path != "" {
			paths = append(paths, p.Path)
		}
	}
	for _, sr := range cfg.ScopedRules {
		if sr.Path != "" {
			paths = append(paths, sr.Path)
		}
	}
	return paths
}

// ── localFetcher ─────────────────────────────────────────────────────────────

// localFetcher handles absolute paths and relative paths starting with ./ or ../
// as well as file:// URLs pointing to a local directory.
type localFetcher struct{}

func (f *localFetcher) CanHandle(source string) bool {
	return strings.HasPrefix(source, "file://") ||
		filepath.IsAbs(source) ||
		strings.HasPrefix(source, "/") || // Unix-style absolute path (also valid on Windows)
		strings.HasPrefix(source, "./") ||
		strings.HasPrefix(source, "../")
}

func (f *localFetcher) FetchAgentsDir(source, destDir string) error {
	root := strings.TrimPrefix(source, "file://")
	agentsDir := filepath.Join(root, ".agents")

	info, err := os.Stat(agentsDir)
	if err != nil {
		return fmt.Errorf("base .agents/ directory not found at %s: %w", agentsDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", agentsDir)
	}

	return copyDirAll(agentsDir, destDir)
}

// copyDirAll recursively copies the contents of src into dst.
func copyDirAll(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// ── httpFetcher ───────────────────────────────────────────────────────────────

// httpFetcher handles http:// and https:// URLs that do not end in ".git".
// It fetches config.json first, then retrieves each file path listed in the config.
type httpFetcher struct{}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func (f *httpFetcher) CanHandle(source string) bool {
	lower := strings.ToLower(source)
	return (strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://")) &&
		!strings.HasSuffix(lower, ".git")
}

func (f *httpFetcher) FetchAgentsDir(source, destDir string) error {
	base := strings.TrimRight(source, "/")

	// Fetch config.json
	configData, err := httpGet(base + "/.agents/config.json")
	if err != nil {
		return fmt.Errorf("fetching config.json: %w", err)
	}

	// Write config.json
	if err := os.WriteFile(filepath.Join(destDir, "config.json"), configData, 0o644); err != nil {
		return err
	}

	// Parse config to discover file paths
	var cfg Config
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return fmt.Errorf("parsing base config.json: %w", err)
	}

	// Fetch each referenced file
	for _, relPath := range collectFilePaths(&cfg) {
		fileURL := base + "/" + relPath
		data, err := httpGet(fileURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not fetch %s: %v\n", fileURL, err)
			continue
		}
		localPath := strings.TrimPrefix(relPath, ".agents/")
		fullPath := filepath.Join(destDir, filepath.FromSlash(localPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// ── gitFetcher ────────────────────────────────────────────────────────────────

// gitFetcher handles git@... SSH URLs and https://....git HTTP URLs.
// It performs a shallow sparse clone to retrieve only the .agents/ directory.
type gitFetcher struct{}

func (f *gitFetcher) CanHandle(source string) bool {
	lower := strings.ToLower(source)
	return strings.HasPrefix(source, "git@") ||
		(strings.HasPrefix(lower, "https://") && strings.HasSuffix(lower, ".git")) ||
		(strings.HasPrefix(lower, "http://") && strings.HasSuffix(lower, ".git"))
}

func (f *gitFetcher) FetchAgentsDir(source, destDir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH (required for git source %q): install git and retry", source)
	}

	// Clone into a temporary directory
	tmpDir, err := os.MkdirTemp("", "ajolote-git-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	clone := exec.Command("git", "clone",
		"--depth", "1",
		"--filter=blob:none",
		"--sparse",
		"--no-tags",
		source, tmpDir)
	clone.Env = env
	if out, err := clone.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s", strings.TrimSpace(string(out)))
	}

	sparse := exec.Command("git", "-C", tmpDir, "sparse-checkout", "set", ".agents")
	sparse.Env = env
	if out, err := sparse.CombinedOutput(); err != nil {
		return fmt.Errorf("git sparse-checkout failed: %s", strings.TrimSpace(string(out)))
	}

	srcAgents := filepath.Join(tmpDir, ".agents")
	if _, err := os.Stat(srcAgents); os.IsNotExist(err) {
		return fmt.Errorf("base repo %q has no .agents/ directory", source)
	}

	return copyDirAll(srcAgents, destDir)
}

// ── ftpFetcher ────────────────────────────────────────────────────────────────

// ftpFetcher handles ftp:// URLs. It uses a minimal built-in FTP client
// (no external dependencies) to retrieve config.json and all referenced files.
// Anonymous login (user "anonymous", password "anonymous@") is used when no
// credentials are supplied in the URL.
type ftpFetcher struct{}

func (f *ftpFetcher) CanHandle(source string) bool {
	return strings.HasPrefix(strings.ToLower(source), "ftp://")
}

func (f *ftpFetcher) FetchAgentsDir(source, destDir string) error {
	u, err := url.Parse(source)
	if err != nil {
		return fmt.Errorf("invalid ftp URL %q: %w", source, err)
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "21"
	}
	root := strings.TrimRight(u.Path, "/")

	user := "anonymous"
	pass := "anonymous@"
	if u.User != nil {
		user = u.User.Username()
		if p, ok := u.User.Password(); ok {
			pass = p
		}
	}

	conn, err := newFTPConn(host+":"+port, user, pass)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Fetch config.json
	configData, err := conn.retr(root + "/.agents/config.json")
	if err != nil {
		return fmt.Errorf("fetching config.json via FTP: %w", err)
	}

	if err := os.WriteFile(filepath.Join(destDir, "config.json"), configData, 0o644); err != nil {
		return err
	}

	var cfg Config
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return fmt.Errorf("parsing base config.json: %w", err)
	}

	for _, relPath := range collectFilePaths(&cfg) {
		remotePath := root + "/" + relPath
		data, err := conn.retr(remotePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not fetch %s via FTP: %v\n", remotePath, err)
			continue
		}
		localPath := strings.TrimPrefix(relPath, ".agents/")
		fullPath := filepath.Join(destDir, filepath.FromSlash(localPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ── minimal FTP client ────────────────────────────────────────────────────────

type ftpConn struct {
	ctrl   net.Conn
	reader *bufio.Reader
}

func newFTPConn(addr, user, pass string) (*ftpConn, error) {
	ctrl, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("ftp connect %s: %w", addr, err)
	}
	c := &ftpConn{ctrl: ctrl, reader: bufio.NewReader(ctrl)}

	// Read welcome banner (220)
	if _, err := c.readResponse(220); err != nil {
		ctrl.Close()
		return nil, err
	}

	// Authenticate
	if err := c.sendCmd("USER "+user, 331, 230); err != nil {
		ctrl.Close()
		return nil, fmt.Errorf("ftp USER: %w", err)
	}
	if err := c.sendCmd("PASS "+pass, 230); err != nil {
		ctrl.Close()
		return nil, fmt.Errorf("ftp PASS: %w", err)
	}

	// Binary mode
	if err := c.sendCmd("TYPE I", 200); err != nil {
		ctrl.Close()
		return nil, fmt.Errorf("ftp TYPE I: %w", err)
	}

	return c, nil
}

func (c *ftpConn) Close() {
	c.sendCmd("QUIT", 221) //nolint:errcheck
	c.ctrl.Close()
}

// retr retrieves the file at remotePath and returns its bytes.
func (c *ftpConn) retr(remotePath string) ([]byte, error) {
	dataAddr, err := c.pasv()
	if err != nil {
		return nil, err
	}

	data, err := net.DialTimeout("tcp", dataAddr, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("ftp data connection: %w", err)
	}
	defer data.Close()

	if err := c.send("RETR " + remotePath); err != nil {
		return nil, err
	}
	if _, err := c.readResponse(150); err != nil {
		return nil, fmt.Errorf("ftp RETR %s: %w", remotePath, err)
	}

	buf, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("ftp reading data: %w", err)
	}

	if _, err := c.readResponse(226); err != nil {
		return nil, fmt.Errorf("ftp transfer complete: %w", err)
	}
	return buf, nil
}

// pasv issues PASV and returns the data connection address.
func (c *ftpConn) pasv() (string, error) {
	if err := c.send("PASV"); err != nil {
		return "", err
	}
	line, err := c.readResponse(227)
	if err != nil {
		return "", fmt.Errorf("ftp PASV: %w", err)
	}
	return parsePASV(line)
}

// parsePASV extracts the host:port from a 227 response line.
// e.g. "227 Entering Passive Mode (192,168,1,1,10,20)"
var pasvRe = regexp.MustCompile(`\((\d+),(\d+),(\d+),(\d+),(\d+),(\d+)\)`)

func parsePASV(line string) (string, error) {
	m := pasvRe.FindStringSubmatch(line)
	if m == nil {
		return "", fmt.Errorf("ftp: cannot parse PASV response: %q", line)
	}
	nums := make([]int, 6)
	for i := range nums {
		nums[i], _ = strconv.Atoi(m[i+1])
	}
	host := fmt.Sprintf("%d.%d.%d.%d", nums[0], nums[1], nums[2], nums[3])
	port := nums[4]*256 + nums[5]
	return fmt.Sprintf("%s:%d", host, port), nil
}

func (c *ftpConn) send(cmd string) error {
	_, err := fmt.Fprintf(c.ctrl, "%s\r\n", cmd)
	return err
}

// sendCmd sends a command and checks that the response code is one of the expected ones.
func (c *ftpConn) sendCmd(cmd string, expected ...int) error {
	if err := c.send(cmd); err != nil {
		return err
	}
	_, err := c.readResponse(expected...)
	return err
}

// readResponse reads FTP response lines until a non-continuation line, then
// verifies the code is in expected.
func (c *ftpConn) readResponse(expected ...int) (string, error) {
	var last string
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("ftp read: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		last = line
		if len(line) >= 4 && line[3] == ' ' {
			break // end of (possibly multi-line) response
		}
	}
	if len(last) < 3 {
		return last, fmt.Errorf("ftp: short response %q", last)
	}
	code, err := strconv.Atoi(last[:3])
	if err != nil {
		return last, fmt.Errorf("ftp: non-numeric response code in %q", last)
	}
	for _, e := range expected {
		if code == e {
			return last, nil
		}
	}
	return last, fmt.Errorf("ftp: unexpected response %q (expected %v)", last, expected)
}

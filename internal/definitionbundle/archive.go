package definitionbundle

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// ReadFile refuses links and bounds caller-selected local files before allocation.
func ReadFile(name string, limit int64) ([]byte, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("%s: expected regular file of at most %d bytes", name, limit)
	}
	//nolint:gosec // Caller-selected paths are explicit local inputs, never bundle commands.
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s: byte limit exceeded", name)
	}
	return data, nil
}

// ReadArchive verifies a pinned archive before exposing any payload bytes.
// Protocol 1 has no compression, links, directory entries, extensions or extraction.
func (m *Manifest) ReadArchive(data []byte, expectedSHA256 string) (map[string][]byte, error) {
	paths := make([]string, 0, len(m.Files))
	for _, file := range m.Files {
		paths = append(paths, file.Path)
	}
	if err := validatePayloadPaths(paths); err != nil {
		return nil, err
	}
	if !digestPattern.MatchString(expectedSHA256) || expectedSHA256 != m.Archive.SHA256 ||
		int64(len(data)) != m.Archive.Size || Digest(data) != expectedSHA256 {
		return nil, fmt.Errorf("archive identity mismatch")
	}
	reader := tar.NewReader(bytes.NewReader(data))
	files := make(map[string][]byte, len(m.Files))
	var encodedSize int64 = 1024
	var offset int64
	for _, file := range m.Files {
		header, err := reader.Next()
		if err != nil {
			return nil, fmt.Errorf("entry %s: %w", file.Path, err)
		}
		if header.Name != file.Path || header.Size != file.Size || header.Typeflag != tar.TypeReg ||
			header.Format != tar.FormatUSTAR || header.Mode != 0o644 || header.Uid != 0 || header.Gid != 0 ||
			header.Uname != "" || header.Gname != "" || header.Linkname != "" ||
			header.ModTime.Unix() != 0 || header.Devmajor != 0 || header.Devminor != 0 ||
			len(header.PAXRecords) != 0 {
			return nil, fmt.Errorf("entry %s: unsupported archive metadata or manifest mismatch", file.Path)
		}
		payload, err := io.ReadAll(io.LimitReader(reader, file.Size+1))
		if err != nil {
			return nil, err
		}
		if int64(len(payload)) != file.Size || Digest(payload) != file.SHA256 {
			return nil, fmt.Errorf("entry %s: content digest/size mismatch", file.Path)
		}
		files[file.Path] = payload
		entrySize := 512 + ((file.Size+511)/512)*512
		paddingStart, paddingEnd := offset+512+file.Size, offset+entrySize
		if paddingEnd > int64(len(data)) ||
			!bytes.Equal(data[paddingStart:paddingEnd], make([]byte, paddingEnd-paddingStart)) {
			return nil, fmt.Errorf("entry %s: nonzero archive padding", file.Path)
		}
		offset += entrySize
		encodedSize += entrySize
	}
	if _, err := reader.Next(); err != io.EOF {
		return nil, fmt.Errorf("unexpected archive entry or trailer")
	}
	if encodedSize != int64(len(data)) || !bytes.Equal(data[len(data)-1024:], make([]byte, 1024)) {
		return nil, fmt.Errorf("archive extensions or trailing bytes are unsupported")
	}
	return files, nil
}

// DefinitionFiles selects the definitions/<name>.yaml payloads of verified
// bundle files, keyed by <name>.yaml. The loaders record that flattened name as
// each rule's definition file, which AuditRequirements matches.
func DefinitionFiles(files map[string][]byte) (map[string][]byte, error) {
	paths := slices.Sorted(maps.Keys(files))
	if err := validatePayloadPaths(paths); err != nil {
		return nil, err
	}
	selected := map[string][]byte{}
	for _, name := range paths {
		if strings.HasPrefix(name, "definitions/") {
			selected[path.Base(name)] = files[name]
		}
	}
	return selected, nil
}

// WriteDefinitions writes only already-verified YAML into a new caller-owned directory.
// The directory must not exist; no archive-selected filesystem paths are used.
func WriteDefinitions(files map[string][]byte, directory string) error {
	selected, err := DefinitionFiles(files)
	if err != nil {
		return err
	}
	if err = os.Mkdir(directory, 0o700); err != nil {
		return err
	}
	for _, name := range slices.Sorted(maps.Keys(selected)) {
		if err = writeNewFile(filepath.Join(directory, name), selected[name]); err != nil {
			return err
		}
	}
	return nil
}

func writeNewFile(target string, data []byte) error {
	//nolint:gosec // Destination is a validated basename within the new caller-owned directory.
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written, err := file.Write(data)
	if err == nil && written != len(data) {
		err = io.ErrShortWrite
	}
	return errors.Join(err, file.Close())
}

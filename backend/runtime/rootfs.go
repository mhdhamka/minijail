package runtime

import (
	"fmt"
	"time"
)

// InitRootfs builds the OverlayFS mount paths and baseline files
func InitRootfs(id, image string) RootfsOverlay {
	shortID := id
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	sanitizedImage := image
	if sanitizedImage == "" {
		sanitizedImage = "alpine:3.19"
	}

	lowerDir := fmt.Sprintf("/var/lib/minijail/images/%s/rootfs", sanitizedImage)
	upperDir := fmt.Sprintf("/var/lib/minijail/containers/%s/diff", shortID)
	workDir := fmt.Sprintf("/var/lib/minijail/containers/%s/work", shortID)
	mergedDir := fmt.Sprintf("/var/lib/minijail/containers/%s/merged", shortID)

	// Default COW initial modifications for a newly booted container
	diffs := []OverlayFileDiff{
		{
			Path:      "/etc/hostname",
			Action:    "CREATED",
			SizeBytes: 14,
			Timestamp: time.Now().Format("15:04:05"),
		},
		{
			Path:      "/etc/hosts",
			Action:    "CREATED",
			SizeBytes: 184,
			Timestamp: time.Now().Format("15:04:05"),
		},
		{
			Path:      "/etc/resolv.conf",
			Action:    "CREATED",
			SizeBytes: 62,
			Timestamp: time.Now().Format("15:04:05"),
		},
	}

	return RootfsOverlay{
		BaseImage:     sanitizedImage,
		LowerDir:      lowerDir,
		UpperDir:      upperDir,
		WorkDir:       workDir,
		MergedDir:     mergedDir,
		ModifiedFiles: diffs,
	}
}

// AddFileDiff registers a new file modification in UpperDir
func (r *RootfsOverlay) AddFileDiff(filePath, action string, sizeBytes int64) {
	r.ModifiedFiles = append(r.ModifiedFiles, OverlayFileDiff{
		Path:      filePath,
		Action:    action,
		SizeBytes: sizeBytes,
		Timestamp: time.Now().Format("15:04:05"),
	})
}

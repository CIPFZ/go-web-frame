package magnet

import "strings"

// Extensions are hints, not an inspection of file contents.
func fileType(ext string) string {
	ext = strings.ToLower(ext)
	for kind, extensions := range map[string]string{
		"video":      "|mp4|mkv|avi|mov|webm|m4v|ts|mts|m2ts|wmv|flv|mpeg|mpg|vob|",
		"audio":      "|flac|mp3|m4a|ogg|wav|aac|alac|opus|wma|aiff|ape|dsf|",
		"image":      "|jpg|jpeg|png|webp|gif|bmp|tif|tiff|avif|heic|",
		"document":   "|pdf|epub|mobi|azw|azw3|cbz|cbr|doc|docx|xls|xlsx|ppt|pptx|txt|md|rtf|csv|odt|",
		"archive":    "|zip|rar|7z|tar|gz|bz2|xz|zst|tgz|tbz2|",
		"disk_image": "|iso|img|nrg|dmg|vhd|vhdx|vmdk|qcow2|bin|",
	} {
		if ext != "" && strings.Contains(extensions, "|"+ext+"|") {
			return kind
		}
	}
	return "other"
}

// Companion artwork/readmes should not replace the main resource type.
// A stable order resolves equal-size ties independently of file order.
func dominantType(sizes map[string]int64, counts map[string]int) string {
	best := ""
	var size int64 = -1
	for _, kind := range []string{"video", "audio", "disk_image", "archive", "document", "other"} {
		if counts[kind] > 0 && sizes[kind] > size {
			best, size = kind, sizes[kind]
		}
	}
	if counts["image"] > 0 && (best == "" || (best == "document" || best == "other") && sizes["image"] > size) {
		return "image"
	}
	if best == "" {
		return "other"
	}
	return best
}

func isImage(ext string) bool {
	// Only supported raster formats can be downloaded as covers.
	return ext == "jpg" || ext == "jpeg" || ext == "png" || ext == "gif" || ext == "webp"
}
func isVideo(ext string) bool { return fileType(ext) == "video" }
func isAudio(ext string) bool { return fileType(ext) == "audio" }

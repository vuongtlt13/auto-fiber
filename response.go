package autofiber

import "github.com/gofiber/fiber/v2"

// FileResponse is a special response type that can send a file to the client
// instead of returning JSON. Any value that implements this interface will be
// detected by AutoFiber and used to send the response directly.
type FileResponse interface {
	// SendFileResponse is responsible for writing the file response to Fiber context.
	SendFileResponse(c *fiber.Ctx) error
}

// DownloadFile is a helper implementation of FileResponse that uses Fiber's Download/SendFile.
//
// Usage in handler:
//
//	func (h *Handler) Download(c *fiber.Ctx) (interface{}, error) {
//	    return autofiber.DownloadFile{
//	        Path:     "./files/report.pdf", // required
//	        FileName: "report.pdf",         // optional suggested filename
//	        Inline:   false,                // false => attachment (Download), true => inline (SendFile)
//	    }, nil
//	}
type DownloadFile struct {
	// Path is the absolute or relative path to the file on disk.
	Path string
	// FileName is an optional suggested name for the downloaded file.
	// If empty and Inline is false, Fiber will use the file name from Path.
	FileName string
	// Inline controls how the browser handles the file:
	// - true  => c.SendFile (typically displayed inline if browser supports it)
	// - false => c.Download (sent as an attachment)
	Inline bool
}

// SendFileResponse implements FileResponse using Fiber's SendFile/Download helpers.
func (d DownloadFile) SendFileResponse(c *fiber.Ctx) error {
	if d.Inline {
		// Display inline using SendFile
		return c.SendFile(d.Path)
	}

	// Send as attachment (Download). Use optional FileName when provided.
	if d.FileName != "" {
		return c.Download(d.Path, d.FileName)
	}
	return c.Download(d.Path)
}



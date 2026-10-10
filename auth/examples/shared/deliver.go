package shared

import "log/slog"

// Deliver stands in for email/SMS: the walkthroughs read codes and links from the server log.
func Deliver(to, msg string) {
	slog.Info("deliver", "to", to, "msg", msg)
}

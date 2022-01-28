package driver

import "context"

type Driver interface {
	Close(c context.Context) (bool, error)                                    // Closes the current session.
	Destroy(c context.Context, sessionId string) (bool, error)                // Destroys a session.
	Gc(c context.Context, maxLifetime int64) (bool, error)                    // Cleans up expired sessions
	Open(c context.Context, path string, name string) (bool, error)           // Re-initialize existing session, or creates a new one.
	Read(c context.Context, sessionId string) (string, error)                 // Reads the session data from the session storage, and returns the results. Called right after the session starts or when session_start() is called.
	Write(c context.Context, sessionId string, value string, ttl int64) error // Writes the session data to the session storage.
	HasSession(c context.Context, sessionId string) (bool, error)             // has session
}

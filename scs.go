/*
Package scs is a session manager for Go 1.7+.

It features:

	* Automatic loading and saving of session data via middleware.
	* Fast and very memory-efficient performance.
	* Choice of PostgreSQL, MySQL, Redis, encrypted cookie and in-memory storage engines. Custom storage engines are also supported.
	* Type-safe and sensible API. Designed to be safe for concurrent use.
	* Supports OWASP good-practices, including absolute and idle session timeouts and easy regeneration of session tokens.

This top-level package is a wrapper for its sub-packages and doesn't actually
contain any code. You probably want to start by looking at the documentation
for the session sub-package.
*/
package scs

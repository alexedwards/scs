package scs

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Status represents the state of the session data during a request cycle.
type Status int

const (
	// Unmodified indicates that the session data hasn't been changed in the
	// current request cycle.
	Unmodified Status = iota

	// Modified indicates that the session data has been changed in the current
	// request cycle.
	Modified

	// Destroyed indicates that the session data has been destroyed in the
	// current request cycle.
	Destroyed
)

type sessionData struct {
	deadline time.Time
	status   Status
	token    string
	values   map[string]interface{}
	mu       sync.Mutex
}

func newSessionData(lifetime time.Duration) *sessionData {
	return &sessionData{
		deadline: time.Now().Add(lifetime).UTC(),
		status:   Unmodified,
		values:   make(map[string]interface{}),
	}
}

// Load retrieves the session data for the given token from the session store,
// and returns a new context.Context containing the session data. If no matching
// token is found then this will create a new session.
//
// Most applications will use the LoadAndSave() middleware and will not need to
// use this method.
func (s *SessionManager) Load(ctx context.Context, token string) (context.Context, error) {
	if _, ok := ctx.Value(s.contextKey).(*sessionData); ok {
		return ctx, nil
	}

	if token == "" {
		return s.addSessionDataToContext(ctx, newSessionData(s.Lifetime)), nil
	}

	b, found, err := s.doStoreFind(ctx, token)
	if err != nil {
		return nil, err
	} else if !found {
		return s.addSessionDataToContext(ctx, newSessionData(s.Lifetime)), nil
	}

	sd := &sessionData{
		status: Unmodified,
		token:  token,
	}
	if sd.deadline, sd.values, err = s.Codec.Decode(b); err != nil {
		return nil, err
	}

	// Mark the session data as modified if an idle timeout is being used. This
	// will force the session data to be re-committed to the session store with
	// a new expiry time.
	if s.IdleTimeout > 0 {
		sd.status = Modified
	}

	return s.addSessionDataToContext(ctx, sd), nil
}

// Commit saves the session data to the session store and returns the session
// token and expiry time.
//
// Most applications will use the LoadAndSave() middleware and will not need to
// use this method.
func (s *SessionManager) Commit(ctx context.Context) (string, time.Time, error) {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.token == "" {
		var err error
		if sd.token, err = generateToken(); err != nil {
			return "", time.Time{}, err
		}
	}

	b, err := s.Codec.Encode(sd.deadline, sd.values)
	if err != nil {
		return "", time.Time{}, err
	}

	expiry := sd.deadline
	if s.IdleTimeout > 0 {
		ie := time.Now().Add(s.IdleTimeout).UTC()
		if ie.Before(expiry) {
			expiry = ie
		}
	}

	if err := s.doStoreCommit(ctx, sd.token, b, expiry); err != nil {
		return "", time.Time{}, err
	}

	return sd.token, expiry, nil
}

// Destroy deletes the session data from the session store and sets the session
// status to Destroyed. Any further operations in the same request cycle will
// result in a new session being created.
func (s *SessionManager) Destroy(ctx context.Context) error {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	err := s.doStoreDelete(ctx, sd.token)
	if err != nil {
		return err
	}

	sd.status = Destroyed

	// Reset everything else to defaults.
	sd.token = ""
	sd.deadline = time.Now().Add(s.Lifetime).UTC()
	for key := range sd.values {
		delete(sd.values, key)
	}

	return nil
}

// Put adds a key and corresponding value to the session data. Any existing
// value for the key will be replaced. The session data status will be set to
// Modified.
func (s *SessionManager) Put(ctx context.Context, key string, val interface{}) {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	sd.values[key] = val
	sd.status = Modified
	sd.mu.Unlock()
}

// Get returns the value for a given key from the session data. The return
// value has the type interface{} so will usually need to be type asserted
// before you can use it. For example:
//
//	foo, ok := session.Get(r, "foo").(string)
//	if !ok {
//		return errors.New("type assertion to string failed")
//	}
//
// Also see the GetString(), GetInt(), GetBytes() and other helper methods which
// wrap the type conversion for common types.
func (s *SessionManager) Get(ctx context.Context, key string) interface{} {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	return sd.values[key]
}

// Pop acts like a one-time Get. It returns the value for a given key from the
// session data and deletes the key and value from the session data. The
// session data status will be set to Modified. The return value has the type
// interface{} so will usually need to be type asserted before you can use it.
func (s *SessionManager) Pop(ctx context.Context, key string) interface{} {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	val, exists := sd.values[key]
	if !exists {
		return nil
	}
	delete(sd.values, key)
	sd.status = Modified

	return val
}

// Remove deletes the given key and corresponding value from the session data.
// The session data status will be set to Modified. If the key is not present
// this operation is a no-op.
func (s *SessionManager) Remove(ctx context.Context, key string) {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	_, exists := sd.values[key]
	if !exists {
		return
	}

	delete(sd.values, key)
	sd.status = Modified
}

// Clear removes all data for the current session. The session token and
// lifetime are unaffected. If there is no data in the current session this is
// a no-op.
func (s *SessionManager) Clear(ctx context.Context) error {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	if len(sd.values) == 0 {
		return nil
	}

	for key := range sd.values {
		delete(sd.values, key)
	}
	sd.status = Modified
	return nil
}

// Exists returns true if the given key is present in the session data.
func (s *SessionManager) Exists(ctx context.Context, key string) bool {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	_, exists := sd.values[key]
	sd.mu.Unlock()

	return exists
}

// Keys returns a slice of all key names present in the session data, sorted
// alphabetically. If the data contains no data then an empty slice will be
// returned.
func (s *SessionManager) Keys(ctx context.Context) []string {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	keys := make([]string, len(sd.values))
	i := 0
	for key := range sd.values {
		keys[i] = key
		i++
	}
	sd.mu.Unlock()

	sort.Strings(keys)
	return keys
}

// RenewToken updates the session data to have a new session token while
// retaining the current session data. The session lifetime is also reset and
// the session data status will be set to Modified.
//
// The old session token and accompanying data are deleted from the session store.
//
// To mitigate the risk of session fixation attacks, it's important that you call
// RenewToken before making any changes to privilege levels (e.g. login and
// logout operations). See https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md#renew-the-session-id-after-any-privilege-level-change
// for additional information.
func (s *SessionManager) RenewToken(ctx context.Context) error {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.token != "" {
		err := s.doStoreDelete(ctx, sd.token)
		if err != nil {
			return err
		}
	}

	newToken, err := generateToken()
	if err != nil {
		return err
	}

	sd.token = newToken
	sd.deadline = time.Now().Add(s.Lifetime).UTC()
	sd.status = Modified

	return nil
}

// MergeSession is used to merge in data from a different session in case strict
// session tokens are lost across an oauth or similar redirect flows. Use Clear()
// if no values of the new session are to be used.
func (s *SessionManager) MergeSession(ctx context.Context, token string) error {
	sd := s.getSessionDataFromContext(ctx)

	b, found, err := s.doStoreFind(ctx, token)
	if err != nil {
		return err
	} else if !found {
		return nil
	}

	deadline, values, err := s.Codec.Decode(b)
	if err != nil {
		return err
	}

	sd.mu.Lock()
	defer sd.mu.Unlock()

	// If it is the same session, nothing needs to be done.
	if sd.token == token {
		return nil
	}

	if deadline.After(sd.deadline) {
		sd.deadline = deadline
	}

	for k, v := range values {
		sd.values[k] = v
	}

	sd.status = Modified
	return s.doStoreDelete(ctx, token)
}

// Status returns the current status of the session data.
func (s *SessionManager) Status(ctx context.Context) Status {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	return sd.status
}

// GetString returns the string value for a given key from the session data.
// The zero value for a string ("") is returned if the key does not exist or the
// value could not be type asserted to a string.
func (s *SessionManager) GetString(ctx context.Context, key string) string {
	val := s.Get(ctx, key)
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

// GetBool returns the bool value for a given key from the session data. The
// zero value for a bool (false) is returned if the key does not exist or the
// value could not be type asserted to a bool.
func (s *SessionManager) GetBool(ctx context.Context, key string) bool {
	val := s.Get(ctx, key)
	b, ok := val.(bool)
	if !ok {
		return false
	}
	return b
}

// GetInt returns the int value for a given key from the session data. The
// zero value for an int (0) is returned if the key does not exist or the
// value could not be type asserted to an int.
func (s *SessionManager) GetInt(ctx context.Context, key string) int {
	val := s.Get(ctx, key)
	i, ok := val.(int)
	if !ok {
		return 0
	}
	return i
}

// GetInt64 returns the int64 value for a given key from the session data. The
// zero value for an int64 (0) is returned if the key does not exist or the
// value could not be type asserted to an int64.
func (s *SessionManager) GetInt64(ctx context.Context, key string) int64 {
	val := s.Get(ctx, key)
	i, ok := val.(int64)
	if !ok {
		return 0
	}
	return i
}

// GetInt32 returns the int value for a given key from the session data. The
// zero value for an int32 (0) is returned if the key does not exist or the
// value could not be type asserted to an int32.
func (s *SessionManager) GetInt32(ctx context.Context, key string) int32 {
	val := s.Get(ctx, key)
	i, ok := val.(int32)
	if !ok {
		return 0
	}
	return i
}

// GetFloat returns the float64 value for a given key from the session data. The
// zero value for an float64 (0) is returned if the key does not exist or the
// value could not be type asserted to a float64.
func (s *SessionManager) GetFloat(ctx context.Context, key string) float64 {
	val := s.Get(ctx, key)
	f, ok := val.(float64)
	if !ok {
		return 0
	}
	return f
}

// GetBytes returns the byte slice ([]byte) value for a given key from the session
// data. The zero value for a slice (nil) is returned if the key does not exist
// or could not be type asserted to []byte.
func (s *SessionManager) GetBytes(ctx context.Context, key string) []byte {
	val := s.Get(ctx, key)
	b, ok := val.([]byte)
	if !ok {
		return nil
	}
	return b
}

// GetTime returns the time.Time value for a given key from the session data. The
// zero value for a time.Time object is returned if the key does not exist or the
// value could not be type asserted to a time.Time. This can be tested with the
// time.IsZero() method.
func (s *SessionManager) GetTime(ctx context.Context, key string) time.Time {
	val := s.Get(ctx, key)
	t, ok := val.(time.Time)
	if !ok {
		return time.Time{}
	}
	return t
}

// PopString returns the string value for a given key and then deletes it from the
// session data. The session data status will be set to Modified. The zero
// value for a string ("") is returned if the key does not exist or the value
// could not be type asserted to a string.
func (s *SessionManager) PopString(ctx context.Context, key string) string {
	val := s.Pop(ctx, key)
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

// PopBool returns the bool value for a given key and then deletes it from the
// session data. The session data status will be set to Modified. The zero
// value for a bool (false) is returned if the key does not exist or the value
// could not be type asserted to a bool.
func (s *SessionManager) PopBool(ctx context.Context, key string) bool {
	val := s.Pop(ctx, key)
	b, ok := val.(bool)
	if !ok {
		return false
	}
	return b
}

// PopInt returns the int value for a given key and then deletes it from the
// session data. The session data status will be set to Modified. The zero
// value for an int (0) is returned if the key does not exist or the value could
// not be type asserted to an int.
func (s *SessionManager) PopInt(ctx context.Context, key string) int {
	val := s.Pop(ctx, key)
	i, ok := val.(int)
	if !ok {
		return 0
	}
	return i
}

// PopFloat returns the float64 value for a given key and then deletes it from the
// session data. The session data status will be set to Modified. The zero
// value for an float64 (0) is returned if the key does not exist or the value
// could not be type asserted to a float64.
func (s *SessionManager) PopFloat(ctx context.Context, key string) float64 {
	val := s.Pop(ctx, key)
	f, ok := val.(float64)
	if !ok {
		return 0
	}
	return f
}

// PopBytes returns the byte slice ([]byte) value for a given key and then
// deletes it from the from the session data. The session data status will be
// set to Modified. The zero value for a slice (nil) is returned if the key does
// not exist or could not be type asserted to []byte.
func (s *SessionManager) PopBytes(ctx context.Context, key string) []byte {
	val := s.Pop(ctx, key)
	b, ok := val.([]byte)
	if !ok {
		return nil
	}
	return b
}

// PopTime returns the time.Time value for a given key and then deletes it from
// the session data. The session data status will be set to Modified. The zero
// value for a time.Time object is returned if the key does not exist or the
// value could not be type asserted to a time.Time.
func (s *SessionManager) PopTime(ctx context.Context, key string) time.Time {
	val := s.Pop(ctx, key)
	t, ok := val.(time.Time)
	if !ok {
		return time.Time{}
	}
	return t
}

// RememberMe controls whether the session cookie is persistent (i.e  whether it
// is retained after a user closes their browser). RememberMe only has an effect
// if you have set SessionManager.Cookie.Persist = false (the default is true) and
// you are using the standard LoadAndSave() middleware.
func (s *SessionManager) RememberMe(ctx context.Context, val bool) {
	s.Put(ctx, "__rememberMe", val)
}

// Iterate retrieves all active (i.e. not expired) sessions from the store and
// executes the provided function fn for each session. If the session store
// being used does not support iteration then Iterate will panic.
func (s *SessionManager) Iterate(ctx context.Context, fn func(context.Context) error) error {
	allSessions, err := s.doStoreAll(ctx)
	if err != nil {
		return err
	}

	for token, b := range allSessions {
		sd := &sessionData{
			status: Unmodified,
			token:  token,
		}

		sd.deadline, sd.values, err = s.Codec.Decode(b)
		if err != nil {
			return err
		}

		ctx = s.addSessionDataToContext(ctx, sd)

		err = fn(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Deadline returns the 'absolute' expiry time for the session. Please note
// that if you are using an idle timeout, it is possible that a session will
// expire due to non-use before the returned deadline.
func (s *SessionManager) Deadline(ctx context.Context) time.Time {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	return sd.deadline
}

// SetDeadline updates the 'absolute' expiry time for the session. Please note
// that if you are using an idle timeout, it is possible that a session will
// expire due to non-use before the set deadline.
func (s *SessionManager) SetDeadline(ctx context.Context, expire time.Time) {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	sd.deadline = expire
	sd.status = Modified
}

// Token returns the session token. Please note that this will return the
// empty string "" if it is called before the session has been committed to
// the store.
func (s *SessionManager) Token(ctx context.Context) string {
	sd := s.getSessionDataFromContext(ctx)

	sd.mu.Lock()
	defer sd.mu.Unlock()

	return sd.token
}

func (s *SessionManager) addSessionDataToContext(ctx context.Context, sd *sessionData) context.Context {
	return context.WithValue(ctx, s.contextKey, sd)
}

func (s *SessionManager) getSessionDataFromContext(ctx context.Context) *sessionData {
	c, ok := ctx.Value(s.contextKey).(*sessionData)
	if !ok {
		panic("scs: no session data in context")
	}
	return c
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

type contextKey string

var (
	contextKeyID      uint64
	contextKeyIDMutex = &sync.Mutex{}
)

func generateContextKey() contextKey {
	contextKeyIDMutex.Lock()
	defer contextKeyIDMutex.Unlock()
	atomic.AddUint64(&contextKeyID, 1)
	return contextKey(fmt.Sprintf("session.%d", contextKeyID))
}

func (s *SessionManager) doStoreDelete(ctx context.Context, token string) (err error) {
	if s.HashTokenInStore {
		token = hashToken(token)
	}
	c, ok := s.Store.(interface {
		DeleteCtx(context.Context, string) error
	})
	if ok {
		return c.DeleteCtx(ctx, token)
	}
	return s.Store.Delete(token)
}

func (s *SessionManager) doStoreFind(ctx context.Context, token string) (b []byte, found bool, err error) {
	if s.HashTokenInStore {
		token = hashToken(token)
	}
	c, ok := s.Store.(interface {
		FindCtx(context.Context, string) ([]byte, bool, error)
	})
	if ok {
		return c.FindCtx(ctx, token)
	}
	return s.Store.Find(token)
}

func (s *SessionManager) doStoreCommit(ctx context.Context, token string, b []byte, expiry time.Time) (err error) {
	if s.HashTokenInStore {
		token = hashToken(token)
	}
	c, ok := s.Store.(interface {
		CommitCtx(context.Context, string, []byte, time.Time) error
	})
	if ok {
		return c.CommitCtx(ctx, token, b, expiry)
	}
	return s.Store.Commit(token, b, expiry)
}

func (s *SessionManager) doStoreAll(ctx context.Context) (map[string][]byte, error) {
	cs, ok := s.Store.(IterableCtxStore)
	if ok {
		return cs.AllCtx(ctx)
	}

	is, ok := s.Store.(IterableStore)
	if ok {
		return is.All()
	}

	panic(fmt.Sprintf("type %T does not support iteration", s.Store))
}

package session

import (
	"fmt"
	"github.com/iizotop/baseweb/db"
	"sync"
	"time"
)

type (
	MemoryProvider struct {
		lock     sync.Mutex
		sessions map[[40]byte]Session
	}

	MemorySession struct {
		Id         [40]byte
		IP         uint32
		UID        uint32
		Perm       uint32
		Name       string
		Email      string
		AccessTime time.Time
		Data       map[string]interface{}
		lock       sync.Mutex
	}
)

func (p *MemoryProvider) Init() {

	p.sessions = map[[40]byte]Session{}
}

func (p *MemoryProvider) Len() int {

	p.lock.Lock()
	defer p.lock.Unlock()

	return len(p.sessions)
}

func (p *MemoryProvider) GC(lifeTime time.Duration) {

	p.lock.Lock()
	defer p.lock.Unlock()

	tm := time.Now()

	for key, memSession := range p.sessions {

		if tm.Sub(memSession.Time()) > lifeTime {

			fmt.Printf("%v\n", memSession)

			db.UpdateSession(key[:])
			delete(p.sessions, key)
		}
	}

	fmt.Println(tm)
}

// ==

func (p *MemoryProvider) SessionInit(keyStr string) Session {

	p.lock.Lock()
	defer p.lock.Unlock()

	memSession := &MemorySession{
		Data:       map[string]interface{}{},
		AccessTime: time.Now(),
	}

	key := [40]byte{}

	copy(key[:], keyStr)
	copy(memSession.Id[:], keyStr)

	p.sessions[key] = memSession

	return memSession
}

func (p *MemoryProvider) SessionRead(keyStr string) Session {

	p.lock.Lock()
	defer p.lock.Unlock()

	key := [40]byte{}
	copy(key[:], keyStr)

	return p.sessions[key]
}

func (p *MemoryProvider) SessionDestroy(keyStr string) {

	p.lock.Lock()
	defer p.lock.Unlock()

	key := [40]byte{}
	copy(key[:], keyStr)

	delete(p.sessions, key)
}

// ==

func (s *MemorySession) Time() (tm time.Time) {

	s.lock.Lock()
	tm = s.AccessTime
	s.lock.Unlock()
	return
}

func (s *MemorySession) Key() string {
	return string(s.Id[:])
}

func (s *MemorySession) SetName(name string) {

	s.lock.Lock()
	s.Name = name
	s.lock.Unlock()
}

func (s *MemorySession) GetName() (name string) {

	s.lock.Lock()
	name = s.Name
	s.lock.Unlock()
	return
}

func (s *MemorySession) SetEmail(email string) {

	s.lock.Lock()
	s.Email = email
	s.lock.Unlock()
}

func (s *MemorySession) GetEmail() (email string) {

	s.lock.Lock()
	email = s.Email
	s.lock.Unlock()
	return
}

func (s *MemorySession) SetUID(uid uint32) {

	s.lock.Lock()
	s.UID = uid
	s.lock.Unlock()
}

func (s *MemorySession) GetUID() (uid uint32) {

	s.lock.Lock()
	uid = s.UID
	s.lock.Unlock()
	return
}

func (s *MemorySession) SetPerm(perm uint32) {

	s.lock.Lock()
	s.Perm = perm
	s.lock.Unlock()
}

func (s *MemorySession) GetPerm() (perm uint32) {

	s.lock.Lock()
	perm = s.Perm
	s.lock.Unlock()
	return
}

func (s *MemorySession) SetTime(tm time.Time) {

	s.lock.Lock()
	s.AccessTime = tm
	s.lock.Unlock()
}

func (s *MemorySession) Clear() {

	s.lock.Lock()
	for name := range s.Data {
		delete(s.Data, name)
	}
	s.lock.Unlock()
}

func (s *MemorySession) SetIp(ip uint32) {

	s.lock.Lock()
	s.IP = ip
	s.lock.Unlock()
}

func (s *MemorySession) GetIp() (ip uint32) {

	s.lock.Lock()
	ip = s.IP
	s.lock.Unlock()
	return
}

func (s *MemorySession) Set(name string, value interface{}) {

	s.lock.Lock()
	s.Data[name] = value
	s.lock.Unlock()
}

func (s *MemorySession) Get(name string) (v interface{}) {

	s.lock.Lock()
	v = s.Data[name]
	s.lock.Unlock()
	return
}

func (s *MemorySession) Delete(name string) {

	s.lock.Lock()
	delete(s.Data, name)
	s.lock.Unlock()
}

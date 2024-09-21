/*
Ciphect, a personal data relay.
Copyright (C) 2024 Naeim Cragwell-Chaudhry

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package api

import "time"

const (
	Ciphect = "ciphect"
	Major   = "0"
	Minor   = "0"
)

type Packet struct {
	Magic  Magic  `json:"magic"`
	Header Header `json:"header"`
	Body   []byte `json:"body"`
}

type Magic [3]string

func NewMagic() Magic { return [3]string{Ciphect, Major, Minor} }

func (magic Magic) String() string       { return magic.Ciphect() + " " + magic.Version() }
func (magic Magic) Ciphect() string      { return magic[0] }
func (magic Magic) Version() string      { return magic.VersionMajor() + "." + magic.VersionMinor() }
func (magic Magic) VersionMajor() string { return magic[1] }
func (magic Magic) VersionMinor() string { return magic[2] }

type Header struct {
	ID         string        `json:"id"`
	Timestamp  time.Time     `json:"timestamp"`
	Expiration time.Duration `json:"expiration"`
	To         []string      `json:"to"`
	From       []string      `json:"from"`
	Type       string        `json:"type"`
}

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

package main

import (
	_ "embed"
	"os"

	"github.com/naeimc/ciphect/internal/cmd"
	"github.com/naeimc/ciphect/internal/cmd/cli"
)

var (
	//go:embed VERSION
	version string

	//go:embed LICENSE
	license string
)

func init() {
	cmd.APPLICATION = os.Args[0]
	cmd.VERSION = version
	cmd.LICENSE = license
}

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}

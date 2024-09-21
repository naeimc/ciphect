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

package cli

import (
	"fmt"
	"runtime"

	"github.com/naeimc/ciphect/internal/cmd"
)

func Main(args []string) int {

	if len(args) == 0 {
		return run()
	}

	switch args[0] {
	case "version":
		return version()
	case "license":
		return license()
	}

	return unknown(args[0])
}

func version() int {
	fmt.Printf("%s v%s %s %s/%s\n", cmd.APPLICATION, cmd.VERSION, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return 0
}

func license() int {
	fmt.Printf(cmd.LICENSE)
	return 0
}

func unknown(command string) int {
	fmt.Printf("%s %s: unknown command\n", cmd.APPLICATION, command)
	return usage()
}

func usage() int {
	fmt.Printf("usage: %s [version|license]\n", cmd.APPLICATION)
	return 0
}

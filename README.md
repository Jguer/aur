[![GoDoc](https://godoc.org/github.com/jguer/aur?status.svg)](https://godoc.org/github.com/jguer/aur)

## aur

Fork of [aur](https://github.com/mikkeloscar/aur) with the following differences:

- Structured client
- Custom http client support
- Request editing

## aur-cli

### Install

```
go get github.com/Jguer/aur/aur-cli
```

Client for simple AUR queries.

**Examples**:

- Search packages in the AUR by name only containing "python37"

```sh
aur-cli -verbose -by name search python37
```

- Search packages in the AUR by name only containing "python37".
  Display results as JSON.

```sh
aur-cli -verbose -json -by name search python37
```

- Find all packages in the AUR maintained by "jguer"

```sh
aur-cli -verbose -by maintainer search jguer
```

- Retrieve information on the package "linux-git"

```sh
aur-cli info linux-git
```

# go wrapper for the AUR JSON API

Wrapper around the json API v5 for AUR found at
http://aur.archlinux.org/rpc

## LICENSE

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

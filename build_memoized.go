// Copyright 2026 Witen Labs
// SPDX-License-Identifier: Apache-2.0

//go:build !coraza.no_memoize

package coraza

import "errors"

func validateBuild() error {
	return errors.New("Witen connector requires build tag coraza.no_memoize: Coraza 3.8.1 caches file-backed patterns by path; see FORK.md")
}

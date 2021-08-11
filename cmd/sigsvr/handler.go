/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package sigsvr

import "github.com/polynetwork/poly/cmd/sigsvr/handlers"

func init() {
	DefCliRpcSvr.RegHandler("createaccount", handlers.CreateAccount)
	DefCliRpcSvr.RegHandler("exportaccount", handlers.ExportAccount)
	DefCliRpcSvr.RegHandler("sigdata", handlers.SigData)
}

/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package vbft

import (
	"fmt"
	"testing"

	"github.com/polynetwork/poly/account"
)

func constructMsg() *blockProposalMsg {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		fmt.Println("GetDefaultAccount error: acc is nil")
		return nil
	}
	msg := constructProposalMsgTest(acc)
	return msg
}
func TestSerializeVbftMsg(t *testing.T) {
	msg := constructMsg()
	_, err := SerializeVbftMsg(msg)
	if err != nil {
		t.Errorf("TestSerializeVbftMsg failed :%v", err)
		return
	}
	t.Logf("TestSerializeVbftMsg succ")
}

func TestDeserializeVbftMsg(t *testing.T) {
	msg := constructMsg()
	data, err := SerializeVbftMsg(msg)
	if err != nil {
		t.Errorf("TestSerializeVbftMsg failed :%v", err)
		return
	}
	_, err = DeserializeVbftMsg(data)
	if err != nil {
		t.Errorf("DeserializeVbftMsg failed :%v", err)
		return
	}
	t.Logf("TestDeserializeVbftMsg succ")
}

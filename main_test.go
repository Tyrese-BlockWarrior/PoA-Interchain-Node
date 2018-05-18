package icn

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMsgHash(t *testing.T) {
	type args struct {
		txHash      common.Hash
		destination common.Address
		value       *big.Int
		data        []byte
		version     uint8
	}
	tests := []struct {
		name string
		args args
		want common.Hash
	}{
		{
			name: "Computes solidity compatible hash",
			args: args{
				common.HexToHash("03c85f1da84d9c6313e0c34bcb5ace945a9b12105988895252b88ce5b769f82b"),
				common.HexToAddress("f17f52151ebef6c7334fad080c5704d77216b732"),
				big.NewInt(100000000),
				[]byte("foo"),
				1,
			},
			want: common.HexToHash("f773eb93f5118ae80006f8fde19c03e7d85465b688ccb3694f62b4b695e4ff35"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MsgHash(tt.args.txHash, tt.args.destination, tt.args.value, tt.args.data, tt.args.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MsgHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

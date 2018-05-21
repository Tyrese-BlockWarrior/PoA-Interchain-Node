package icn

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMsgHash(t *testing.T) {
	type args struct {
		contractAddress common.Address
		txHash          common.Hash
		toAddress       common.Address
		value           *big.Int
		data            []byte
		version         uint8
	}
	tests := []struct {
		name string
		args args
		want common.Hash
	}{
		{
			name: "Computes solidity compatible hash",
			args: args{
				common.HexToAddress("75076e4fbba61f65efb41d64e45cff340b1e518a"),
				common.HexToHash("03c85f1da84d9c6313e0c34bcb5ace945a9b12105988895252b88ce5b769f82b"),
				common.HexToAddress("f17f52151ebef6c7334fad080c5704d77216b732"),
				big.NewInt(100000000),
				[]byte{},
				1,
			},
			want: common.HexToHash("6b0673bcb3726c0f7956ef57a9542ed225bfe74f1d2a75414d198d55e8956da5"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MsgHash(
				tt.args.contractAddress,
				tt.args.txHash,
				tt.args.toAddress,
				tt.args.value,
				tt.args.data,
				tt.args.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MsgHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MsgHash(tt.args.contractAddress, tt.args.txHash, tt.args.toAddress, tt.args.value, tt.args.data, tt.args.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MsgHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSignature(t *testing.T) {
	type args struct {
		sig []byte
	}
	tests := []struct {
		name  string
		args  args
		wantV uint8
		wantR common.Hash
		wantS common.Hash
	}{
		{
			name: "Parses signature correctly",
			args: args{
				sig: common.Hex2Bytes("a27a17b20a8dcc6fedb6196b84624ce3f3961a2423642fe13003a816c383f93205adf64e0805449d18b866991ce19e5439567cd3613ae1775e90fb4a8b0cbc6800"),
			},
			wantV: 27,
			wantR: common.HexToHash("a27a17b20a8dcc6fedb6196b84624ce3f3961a2423642fe13003a816c383f932"),
			wantS: common.HexToHash("05adf64e0805449d18b866991ce19e5439567cd3613ae1775e90fb4a8b0cbc68"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, gotR, gotS := ParseSignature(tt.args.sig)
			if gotV != tt.wantV {
				t.Errorf("ParseSignature() gotV = %v, want %v", gotV, tt.wantV)
			}
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("ParseSignature() gotR = %v, want %v", gotR, tt.wantR)
			}
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("ParseSignature() gotS = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

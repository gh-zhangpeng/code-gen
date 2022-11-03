package mysql_gen

import "testing"

func Test_formatTable2Struct(t *testing.T) {
	type args struct {
		table string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{name: "1", args: args{table: "tblUser"}},
		{name: "2", args: args{table: "tblUser1"}},
		{name: "3", args: args{table: "tblUser_1"}},
		{name: "4", args: args{table: "tbl1User"}},
		{name: "5", args: args{table: "1tblUser"}},
		{name: "6", args: args{table: "tbl_user"}},
		{name: "7", args: args{table: "tbl-U-ser"}},
		{name: "8", args: args{table: "tbl-User-1"}},
		{name: "999", args: args{table: "999"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatTable2Struct(tt.args.table)
		})
	}
}

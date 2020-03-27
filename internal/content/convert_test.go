package convert

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func Test_HandlerFromZipJSON(t *testing.T) {
	tests := []struct {
		filename string
		name     string
		want     []byte
		wantErr  bool
	}{
		{
			filename: "tests/webUsage.zip",
			name:     "test bare zip data",
			want: []byte(`{"one": 1, "two": 2}
`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("HandlerFromZipJSON() error = %v", err)
			}
			got, err := HandlerFromZipJSON(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandlerFromZipJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandlerFromZipJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func Test_HandlerFromBase64ZipJSON(t *testing.T) {
	tests := []struct {
		filename string
		name     string
		want     []byte
		wantErr  bool
	}{
		{
			filename: "tests/webUsage.b64",
			name:     "test zipped data",
			want: []byte(`{"one": 1, "two": 2}
`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("HandlerFromBase64ZipJSON() error = %v", err)
			}
			got, err := HandlerFromBase64ZipJSON(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandlerFromBase64ZipJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandlerFromBase64ZipJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func Test_decodeBase64(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeBase64(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unzipFirstFile(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unzipFirstFile(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("unzipFirstFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unzipFirstFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

package objst

import (
	"testing"

	"github.com/google/uuid"
)

func TestQuery_isValid(t *testing.T) {
	type fields struct {
		params *Metadata
		act    action
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "empty params error",
			fields: fields{
				params: NewMetadata(),
			},
			wantErr: true,
		},
		{
			name: "missing owner but name set",
			fields: fields{
				params: &Metadata{
					data: map[MetaKey]string{
						"name": "something",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wrong id and owner format",
			fields: fields{
				params: &Metadata{
					data: map[MetaKey]string{
						"id": "diasjdkas",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "pass query",
			fields: fields{
				params: &Metadata{
					data: map[MetaKey]string{
						"name":  "/some/path/something.json",
						"owner": uuid.NewString(),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{
				params: tt.fields.params,
				act:    tt.fields.act,
			}
			if err := q.isValid(); (err != nil) != tt.wantErr {
				t.Errorf("Query.isValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package main_test

import (
	"testing"

	main "github.com/bluegopher/go-musthave-metrics-tpl/cmd/sum_1"
	"github.com/stretchr/testify/assert"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{
			name:  "simple test #1",
			value: 3.1,
			want:  3.1,
		},
		{
			name:  "simple test #2",
			value: -3,
			want:  3,
		},
		{
			name:  "simple test #3",
			value: 3,
			want:  3,
		},
		{
			name:  "simple test #4",
			value: -2.000001,
			want:  2.000001,
		},
		{
			name:  "simple test #5",
			value: -0.000000003,
			want:  0.000000003,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := main.Abs(tt.value)
			if got != tt.want {
				t.Errorf("Abs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAbs_New(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{
			name:  "simple test #1",
			value: 3.1,
			want:  3.1,
		},
		{
			name:  "simple test #2",
			value: -3,
			want:  3,
		},
		{
			name:  "simple test #3",
			value: 3,
			want:  3,
		},
		{
			name:  "simple test #4",
			value: -2.000001,
			want:  2.000001,
		},
		{
			name:  "simple test #5",
			value: -0.000000003,
			want:  0.000000003,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, main.Abs(tt.value))
		})
	}
}

func TestUser_FullName(t *testing.T) {
	type fields struct {
		FirstName string
		LastName  string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test #1",
			fields: fields{
				FirstName: "Misha",
				LastName:  "Popov",
			},
			want: "Misha Popov",
		},

		{
			name: "test #2",
			fields: fields{
				FirstName: "Ivan",
				LastName:  "Ivanov",
			},
			want: "Ivan Ivanov",
		},

		{
			name: "test #3",
			fields: fields{
				FirstName: "Vasiliy",
				LastName:  "Vasilev",
			},
			want: "Vasiliy Vasilev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			u := main.User{
				FirstName: tt.fields.FirstName,
				LastName:  tt.fields.LastName,
			}

			if gg := u.FullName(); tt.want != gg {
				t.Errorf("FullName() = %v, want %v", gg, tt.want)
			}
		})
	}
}

func TestUser_FullName_New(t *testing.T) {
	type fields struct {
		FirstName string
		LastName  string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test #1",
			fields: fields{
				FirstName: "Misha",
				LastName:  "Popov",
			},
			want: "Misha Popov",
		},

		{
			name: "test #2",
			fields: fields{
				FirstName: "Ivan",
				LastName:  "Ivanov",
			},
			want: "Ivan Ivanov",
		},

		{
			name: "test #3",
			fields: fields{
				FirstName: "Vasiliy",
				LastName:  "Vasilev",
			},
			want: "Vasiliy Vasilev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			u := main.User{
				FirstName: tt.fields.FirstName,
				LastName:  tt.fields.LastName,
			}
			assert.Equal(t, tt.want, u.FullName())
		})
	}
}

func TestFamily_AddNew(t *testing.T) {
	type newPerson struct {
		r main.Relationship
		p main.Person
	}

	tests := []struct {
		name           string
		existedMembers map[main.Relationship]main.Person
		newPerson      newPerson
		wantErr        bool
	}{
		{
			name: "test #1",
			existedMembers: map[main.Relationship]main.Person{
				main.Mother: {
					FirstName: "Maria",
					LastName:  "Popova",
					Age:       36,
				},
			},
			newPerson: newPerson{
				r: main.Father,
				p: main.Person{
					FirstName: "Misha",
					LastName:  "Popov",
					Age:       42,
				},
			},
			wantErr: false,
		},
		{
			name: "test #2",
			existedMembers: map[main.Relationship]main.Person{
				main.Father: {
					FirstName: "Misha",
					LastName:  "Popov",
					Age:       42,
				},
			},
			newPerson: newPerson{
				r: main.Father,
				p: main.Person{
					FirstName: "Ivan",
					LastName:  "Ivanov",
					Age:       44,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//var f main.Family

			f := &main.Family{
				Members: tt.existedMembers,
			}

			gotErr := f.AddNew(tt.newPerson.r, tt.newPerson.p)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AddNew() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AddNew() succeeded unexpectedly")
			}
		})
	}
}

func TestFamily_AddNew_new(t *testing.T) {
	type newPerson struct {
		r main.Relationship
		p main.Person
	}

	tests := []struct {
		name           string
		existedMembers map[main.Relationship]main.Person
		newPerson      newPerson
		wantErr        bool
	}{
		{
			name: "test #1",
			existedMembers: map[main.Relationship]main.Person{
				main.Mother: {
					FirstName: "Maria",
					LastName:  "Popova",
					Age:       36,
				},
			},
			newPerson: newPerson{
				r: main.Father,
				p: main.Person{
					FirstName: "Misha",
					LastName:  "Popov",
					Age:       42,
				},
			},
			wantErr: false,
		},
		{
			name: "test #2",
			existedMembers: map[main.Relationship]main.Person{
				main.Father: {
					FirstName: "Misha",
					LastName:  "Popov",
					Age:       42,
				},
			},
			newPerson: newPerson{
				r: main.Father,
				p: main.Person{
					FirstName: "Ivan",
					LastName:  "Ivanov",
					Age:       44,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &main.Family{
				Members: tt.existedMembers,
			}

			gotErr := f.AddNew(tt.newPerson.r, tt.newPerson.p)

			if !tt.wantErr {
				assert.NoError(t, gotErr)
				assert.Contains(t, f.Members, tt.newPerson.r)
				return
			}
			assert.Error(t, gotErr)
		})
	}
}

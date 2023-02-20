package list

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_insert(t *testing.T) {
	type args struct {
		l *list
		v int
	}
	tests := []struct {
		name string
		args args
		want *list
	}{
		{
			name: "insert with gap",
			args: args{
				l: &list{},
				v: 2,
			},
			want: func() *list {
				l := &list{Min: 0, Max: 0}
				n := &list{Min: 2, Max: 2}
				l.Next = n
				n.Prev = l
				return l
			}(),
		},
		{
			name: "close the gap",
			args: args{
				l: func() *list {
					l := &list{Min: 0, Max: 0}
					n := &list{Min: 2, Max: 2}
					l.Next = n
					n.Prev = l
					return l
				}(),
				v: 1,
			},
			want: &list{Min: 0, Max: 2},
		},
		{
			name: "close the gap 2",
			args: args{
				l: func() *list {
					l := &list{Min: 0, Max: 0}
					n := &list{Min: 2, Max: 2}
					nn := &list{Min: 4, Max: 4}
					l.Next = n
					n.Prev = l
					n.Next = nn
					nn.Prev = n
					return l
				}(),
				v: 3,
			},
			want: func() *list {
				l := &list{Min: 0, Max: 0}
				n := &list{Min: 2, Max: 4}
				l.Next = n
				n.Prev = l
				return l
			}(),
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := insert(tt.args.l, tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInserts(t *testing.T) {
	l := &list{}
	for i := 0; i < 100_000; i = i + 2 {
		l = insert(l, i)
	}
	for i := 99_999; i >= 0; i = i - 2 {
		l = insert(l, i)
	}

	assert.Equal(t, &list{Min: 0, Max: 99_999}, l)
}

func TestRandomInserts(t *testing.T) {
	n := &list{}
	max := 100_000
	var bla []int
	for i := 0; i < max; i++ {
		bla = append(bla, i)
	}
	for j := 0; j < max; j++ {
		i := rand.Intn(len(bla))
		n = insert(n, bla[i])
		bla = append(bla[:i], bla[i+1:]...)
	}

	assert.Equal(t, &list{Min: 0, Max: max - 1}, n)
}

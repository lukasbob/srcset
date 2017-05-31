package srcset

import (
	"reflect"
	"testing"
)

func fl(x float64) *float64 {
	return &x
}

func i(x int64) *int64 {
	return &x
}

func Test_parse(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want SourceSet
	}{
		{
			name: "URL only",
			args: args{"logo-printer-friendly.svg"},
			want: SourceSet{
				ImageSource{URL: "logo-printer-friendly.svg"},
			},
		},
		{
			name: "Parse URL & density",
			args: args{"image-1x.png 1x, image-2x.png 2x, image-3x.png 3x, image-4x.png 4x"},
			want: SourceSet{
				ImageSource{URL: "image-1x.png", Density: fl(1)},
				ImageSource{URL: "image-2x.png", Density: fl(2)},
				ImageSource{URL: "image-3x.png", Density: fl(3)},
				ImageSource{URL: "image-4x.png", Density: fl(4)},
			},
		},
		{
			name: "Parse URL & width - with line break whitespace",
			args: args{`elva-fairy-320w.jpg 320w,
			            elva-fairy-480w.jpg 480w,
			            elva-fairy-800w.jpg 800w`},
			want: SourceSet{
				ImageSource{URL: "elva-fairy-320w.jpg", Width: i(320)},
				ImageSource{URL: "elva-fairy-480w.jpg", Width: i(480)},
				ImageSource{URL: "elva-fairy-800w.jpg", Width: i(800)},
			},
		},
		{
			name: "Parse URL & height - with line break whitespace",
			args: args{`elva-fairy-320h.jpg 320h,
			            elva-fairy-480h.jpg 480h,
			            elva-fairy-800h.jpg 800h`},
			want: SourceSet{
				ImageSource{URL: "elva-fairy-320h.jpg", Height: i(320)},
				ImageSource{URL: "elva-fairy-480h.jpg", Height: i(480)},
				ImageSource{URL: "elva-fairy-800h.jpg", Height: i(800)},
			},
		},
		{
			name: "Invalid: Multiple densities",
			args: args{"test.png 1x 2x"},
			want: SourceSet{},
		},
		{
			name: "Invalid: Density and width",
			args: args{"test.png 1x 200w"},
			want: SourceSet{},
		},
		{
			name: "Invalid: negative width",
			args: args{"test.png -100w"},
			want: SourceSet{},
		},
		{
			name: "Invalid: zero width",
			args: args{"test.png 0w"},
			want: SourceSet{},
		},
		{
			name: "Invalid: None-number width",
			args: args{"test.png f55w"},
			want: SourceSet{},
		},
		{
			name: "Invalid: negative height",
			args: args{"test.png -100h"},
			want: SourceSet{},
		},
		{
			name: "Invalid: zero height",
			args: args{"test.png 0h"},
			want: SourceSet{},
		},
		{
			name: "Invalid: multiple heights",
			args: args{"test.png 124h 234h"},
			want: SourceSet{},
		},
		{
			name: "Invalid: negative density",
			args: args{"test.png -1.3x"},
			want: SourceSet{},
		},

		{
			name: "Super funky",
			args: args{"data:,a ( , data:,b 1x, ), data:,c"},
			want: SourceSet{
				ImageSource{URL: "data:,c"},
			},
		},
	}

	for _, tt := range tests {
		if got := Parse(tt.args.input); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Parse() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

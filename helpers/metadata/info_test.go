package metadata_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/sdk"
)

func TestMergeItems(t *testing.T) {
	t.Parallel()

	testUUID := uuid.MustParse("35802320-e560-4fb3-a380-0dd49da326c0")
	testTime := time.Now()
	testTimeUpdated := testTime.Add(time.Hour)

	type args struct {
		dst sdk.Item
		src sdk.Item
	}

	tests := []struct {
		name string
		args args
		want sdk.Item
	}{
		{
			name: "Movie",
			args: args{
				dst: sdk.Movie{
					ItemInfo: &sdk.ItemInfo{
						ID:            1,
						UUID:          testUUID,
						Title:         "Test Movie",
						SortTitle:     "Test Movie",
						OriginalTitle: "Test Movie",
						ReleaseDate:   testTime,
						Thumb: sdk.Posters{
							Items: []sdk.ItemImage{
								{
									External:  true,
									URL:       "https://image.tmdb.org/t/p/w500/test.jpg",
									Provider:  "tv.meteorae.agents.themoviedb",
									Media:     "metadata://tv.meteorae.agents.themoviedb_030dc1f936c3415aff3f3357163515190d347a28e758e1f717d17bae453541c9",
									SortOrder: 0,
								},
							},
						},
						Parts: []string{
							"Test Part 1",
						},
						CreatedAt: testTime,
						UpdatedAt: testTime,
						DeletedAt: testTime,
					},
				},
				src: sdk.Movie{
					ItemInfo: &sdk.ItemInfo{
						Title:         "Test Movie 2",
						SortTitle:     "Test Movie 2",
						OriginalTitle: "Test Movie 2",
						ReleaseDate:   testTimeUpdated,
						Thumb: sdk.Posters{
							Items: []sdk.ItemImage{
								{
									External:  true,
									URL:       "https://assets.fanart.tv/fanart/movies/71138/movieposter/female-prisoner-701-scorpion-5aa0947cc7e28.jpg",
									Provider:  "tv.meteorae.agents.fanarttv",
									Media:     "metadata://tv.meteorae.agents.fanarttv_c12cbbdba37f89342504115f3ea3c720cf6fa29827b4bdd7a85fe1a284ef53a9",
									SortOrder: 0,
								},
							},
						},
						Parts: []string{
							"Test Part 1",
						},
						CreatedAt: testTime,
						UpdatedAt: testTime,
						DeletedAt: testTime,
					},
				},
			},
			want: sdk.Movie{
				ItemInfo: &sdk.ItemInfo{
					ID:            1,
					UUID:          testUUID,
					Title:         "Test Movie 2",
					SortTitle:     "Test Movie 2",
					OriginalTitle: "Test Movie 2",
					ReleaseDate:   testTimeUpdated,
					Thumb: sdk.Posters{
						Items: []sdk.ItemImage{
							{
								External:  true,
								URL:       "https://assets.fanart.tv/fanart/movies/71138/movieposter/female-prisoner-701-scorpion-5aa0947cc7e28.jpg",
								Provider:  "tv.meteorae.agents.fanarttv",
								Media:     "metadata://tv.meteorae.agents.fanarttv_c12cbbdba37f89342504115f3ea3c720cf6fa29827b4bdd7a85fe1a284ef53a9",
								SortOrder: 0,
							},
							{
								External:  true,
								URL:       "https://image.tmdb.org/t/p/w500/test.jpg",
								Provider:  "tv.meteorae.agents.themoviedb",
								Media:     "metadata://tv.meteorae.agents.themoviedb_030dc1f936c3415aff3f3357163515190d347a28e758e1f717d17bae453541c9",
								SortOrder: 1,
							},
						},
					},
					Parts: []string{
						"Test Part 1",
					},
					CreatedAt: testTime,
					UpdatedAt: testTime,
					DeletedAt: testTime,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := metadata.MergeItems(tt.args.dst, tt.args.src); !reflect.DeepEqual(got, tt.want) {
				gotJson, _ := json.Marshal(got)
				wantJson, _ := json.Marshal(tt.want)
				t.Errorf("MergeItems() = %s, want %s", gotJson, wantJson)
			}
		})
	}
}

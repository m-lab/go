package bqfake_test

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/bigquery/bqiface"
	"github.com/m-lab/go/cloud/bqfake"
	"github.com/m-lab/go/internal/frombigquery"
	"google.golang.org/api/iterator"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// This fails to compile if Dataset does not satisfy the interface.
func assertDataset(ds bqfake.Dataset) {
	func(cc bqiface.Dataset) {}(ds)
}

// This fails to compile if Client does not satisfy the interface.
func assertClient(c bqfake.Client) {
	func(cc bqiface.Client) {}(c)
}

func TestDryRunClient(t *testing.T) {
	os.Setenv("VERBOSE_CLIENT", "true")
	c, ct := bqfake.DryRunClient()

	r, err := c.Get("http://foobar.com/request?foo=bar")
	defer r.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != http.StatusOK {
		t.Fatal("wrong status code", r.StatusCode)
	}
	if ct.Count() != 1 {
		t.Error("Didn't see the Get")
	}
	if len(ct.Requests()) != 1 {
		t.Error("Wrong number of requests")
	}
}

func TestNewClientErr(t *testing.T) {
	ctx := context.Background()
	opts := []option.ClientOption{option.WithAPIKey("asdf"), option.WithoutAuthentication()}
	c, err := bqfake.NewClient(ctx, "fakeProject", opts...)
	if err == nil {
		c.Close()
		t.Fatal("Should return dial error")
	} else if !strings.Contains(err.Error(), "dialing") {
		t.Fatal("Should return dial error:", err.Error())
	}
}

func TestBadDataset(t *testing.T) {
	ds := bqfake.Dataset{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	ds.Table("Foobar")
}

func TestUninitializedTable(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}
	ds := c.Dataset("fakeDataset")

	tbl := ds.Table("DedupTest")
	meta, err := tbl.Metadata(ctx)
	if err == nil {
		t.Error("should return an error")
	} else if err.Error() != "Error 404: Not found: Table fakeProject:fakeDataset.DedupTest, notFound" {
		t.Error("wrong error", err)
	}
	if meta != nil {
		t.Error("meta should be nil")
	}

	// Improperly initialized Table
	tbl = bqfake.Table{}
	meta, err = tbl.Metadata(ctx)
	if err == nil {
		t.Error("Should return an initialization error")
	} else if err.Error() != "Table object incorrectly initialized" {
		t.Error("Incorrect error:", err)
	}

	meta = &bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
	err = tbl.Create(ctx, meta)
	if err == nil {
		t.Error("Should return an initialization error")
	} else if err.Error() != "Table object incorrectly initialized" {
		t.Error("Incorrect error:", err)
	}
}

func TestTable(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}
	ds := c.Dataset("etl")

	{
		meta := bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
		meta.TimePartitioning = &bigquery.TimePartitioning{Expiration: 0 * time.Second}
		tbl := ds.Table("DedupTest")
		err := tbl.Create(ctx, &meta)
		if err != nil {
			t.Fatal(err)
		}
	}

	tbl := ds.Table("DedupTest")
	fqn := tbl.FullyQualifiedName()
	if fqn != "fakeProject:etl.DedupTest" {
		t.Error("Got", fqn)
	}

	if tbl.TableID() != "DedupTest" {
		t.Error("Expected TableID() = DedupTest, got", tbl.TableID())
	}
}

func createTable(ctx context.Context, ds bqiface.Dataset, name string) error {
	meta := bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
	meta.TimePartitioning = &bigquery.TimePartitioning{Expiration: 0 * time.Second}
	tbl := ds.Table("DedupTest")
	return tbl.Create(ctx, &meta)
}

func TestTableMetadata(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}
	ds := c.Dataset("etl")

	// TODO: Test whether changes to table can be seen in existing table objects.
	err = createTable(ctx, ds, "DedupTest")
	if err != nil {
		t.Fatal(err)
	}

	tbl := ds.Table("DedupTest")
	log.Println(tbl)
	meta, err := tbl.Metadata(ctx)
	if err != nil {
		t.Error(err)
	} else if meta == nil {
		t.Error("Meta should not be nil")
	}

	// Try to create the table again - should get an error
	err = tbl.Create(ctx, &bigquery.TableMetadata{})
	if err == nil {
		t.Fatal("Should throw error")
	} else {
		log.Println(err)
	}
}

func TestQuery(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}

	q := c.Query("foobar")
	q.SetQueryConfig(bqiface.QueryConfig{})
	j, err := q.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}
	it, err := q.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = it.Next(struct{}{})
	if err != iterator.Done {
		t.Fatal("Expected iterator.Done, got", err)
	}
}

type foobar struct {
	A int
	B string
}

func TestUploader(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}

	ds := c.Dataset("fakeDataset")

	// TODO: Test whether changes to table can be seen in existing table objects.
	err = createTable(ctx, ds, "DedupTest")
	if err != nil {
		t.Fatal(err)
	}

	tbl := ds.Table("fakeTable")

	err = tbl.Uploader().Put(ctx, foobar{A: 123, B: "foobar"})
	if err != nil {
		t.Fatal(err)
	}

	//	err = tbl.Uploader().Put(ctx, map[string]bigquery.Value{"foobar": 1234})
	//	if err != nil {
	//		t.Fatal(err)
	//	}

	rows := tbl.Uploader().(*frombigquery.FakeUploader).Rows
	for _, r := range rows {
		log.Printf("%+v\n", r.Row)
	}

}

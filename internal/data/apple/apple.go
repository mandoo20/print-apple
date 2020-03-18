package apple

import (
	"context"
	"log"

	appleEntity "print-apple/internal/entity/apple"
	// "print-apple/pkg/errors"
	"print-apple/pkg/errors"
	firebaseclient "print-apple/pkg/firebaseClient"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	// "github.com/jmoiron/sqlx"
)

type (
	// Data ...
	Data struct {
		fb *firestore.Client
	}

	// statement ...
	statement struct {
		key   string
		query string
	}
)

const (
	getAllUsers  = "GetAllUsers"
	qGetAllUsers = "SELECT * FROM user_test"
)

var (
	readStmt = []statement{
		{getAllUsers, qGetAllUsers},
	}
)

// New ...
// db *sqlx.DB
func New(fb *firebaseclient.Client) Data {
	d := Data{
		// db: db,
		fb: fb.Client,
	}

	// d.initStmt()
	return d
}

// GetPrintApple ...
func (d Data) GetPrintApple(ctx context.Context) ([]appleEntity.Apple, error) {
	var (
		appleFirebase []appleEntity.Apple
		err           error
	)
	// test := d.fb.Collection("user_test")
	iter := d.fb.Collection("PrintApple").Documents(ctx)
	for {
		var apple appleEntity.Apple
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		log.Println(doc)
		err = doc.DataTo(&apple)
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(apple)
		appleFirebase = append(appleFirebase, apple)
	}
	return appleFirebase, err
}

//Get and set data

//UpdateStorage ...
func (d Data) UpdateStorage(ctx context.Context, TransFH string) error {
	_, err := d.fb.Collection("PrintApple").Doc(TransFH).Update(ctx, []firestore.Update{{
		Path: "printCount", Value: 1}, {Path: "printed", Value: "Y"}})
	doc, err := d.fb.Collection("PrintApple").Doc(TransFH).Get(ctx)
	appleValidate := doc.Data()
	if appleValidate != nil {

	} else if appleValidate == nil {
		return errors.Wrap(err, "Data Not Exist")
	}
	log.Println(appleValidate)
	_, err = d.fb.Collection("PrintAppleStorage").Doc(TransFH).Set(ctx, appleValidate)
	return err
}

// DeleteAndUpdateStorage ...
func (d Data) DeleteAndUpdateStorage(ctx context.Context, TransFH string) error {
	doc, err := d.fb.Collection("PrintApple").Doc(TransFH).Get(ctx)
	appleValidate := doc.Data()
	if appleValidate == nil {
		return errors.Wrap(err, "Data Not Exist")
	}
	_, err = d.fb.Collection("PrintApple").Doc(TransFH).Delete(ctx)
	return err
}

//Insert ...
func (d Data) Insert(ctx context.Context, apple appleEntity.Apple) error {
	_, err := d.fb.Collection("PrintApple").Doc(apple.TransFH).Set(ctx, apple)

	return err
}

// GetPrintPage ...
func (d Data) GetPrintPage(ctx context.Context, page int, length int) ([]appleEntity.Apple, error) {
	var (
		apple   appleEntity.Apple
		apples  []appleEntity.Apple
		iter    *firestore.DocumentIterator
		lastDoc *firestore.DocumentSnapshot
		err     error
	)

	if page == 1 {
		// Kalau page 1 ambil data langsung dari query
		iter = d.fb.Collection("PrintApple").Limit(length).Documents(ctx)
	} else {
		// Kalau page > 1 ambil data sampai page sebelumnya
		previous := d.fb.Collection("PrintApple").Limit((page - 1) * length).Documents(ctx)
		docs, err := previous.GetAll()
		if err != nil {
			return nil, err
		}
		// Ambil doc terakhir
		lastDoc = docs[len(docs)-1]
		// Query mulai setelah doc terakhir
		iter = d.fb.Collection("PrintApple").StartAfter(lastDoc).Limit(length).Documents(ctx)
	}

	// Looping documents
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return apples, errors.Wrap(err, "[DATA][GetUserPage] Failed to iterate Document!")
		}
		err = doc.DataTo(&apple)
		if err != nil {
			return apples, errors.Wrap(err, "[DATA][GetUserPage] Failed to Populate Struct!")
		}
		apples = append(apples, apple)
	}
	return apples, err
}
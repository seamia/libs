package snapshot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	prn "github.com/seamia/libs/printer"
	"github.com/seamia/libs/zip"
)

type msi = map[string]interface{}
type m2s = map[string]string
type msm2s = map[string]m2s

const (
	folder                 = "state/"
	maxAllowedItemCount    = 1000 // a guard against processing huge tables (limiting by the record count)
	compressPersistedState = true
)

var (
	substitutes = map[string]string{
		"USER": "",
	}
)

type TableDesc struct {
	Name           string                       `json:"name"`
	TableArn       string                       `json:"arn"`
	KeySchema      []*dynamodb.KeySchemaElement `json:"schema"`
	TableSizeBytes int64                        `json:"bytes"`
	ItemCount      int64                        `json:"items"`
	Data           msm2s                        `json:"data"`
}

type Snapshot struct {
	Tables  map[string]TableDesc `json:"tables"`
	Comment string               `json:"comment,omitempty"`
}

func New(svc *dynamodb.DynamoDB, comment string) (*Snapshot, error) {
	local := &Snapshot{}
	if err := local.Snap(svc, comment); err != nil {
		return nil, err
	}
	return local, nil
}

func Load(name string) (*Snapshot, error) {
	local := &Snapshot{}
	if err := local.Load(name); err != nil {
		return nil, err
	}
	return local, nil
}

func (snap *Snapshot) Snap(svc *dynamodb.DynamoDB, comment string) error {
	input := &dynamodb.ListTablesInput{}
	result, err := svc.ListTables(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}

	snap.Tables = make(map[string]TableDesc)
	fmt.Println("Available tables:", result)
	for _, one := range result.TableNames {
		if err = lookat(svc, *one, snap); err != nil {
			return err
		}
	}

	snap.Comment = comment
	return err
}

func (snap *Snapshot) Save(name string) error {

	raw, err := json.Marshal(snap)
	if err == nil {
		if compressPersistedState {
			raw = zip.Compress(raw)
		}
		err = ioutil.WriteFile(folder+name+".state", raw, 0644)
		return err
	} else {
		return err
	}
}

func (snap *Snapshot) Load(name string) error {

	filename := folder + name + ".state"
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// potential de-compression
	raw, err = zip.Decompress(raw)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(raw, snap); err != nil {
		return err
	}

	return nil
}

func (snap *Snapshot) CompareWith(other *Snapshot, printer prn.Printer) {
	if printer == nil {
		printer = prn.Stdout // default to stdout
	}

	if snap == nil || other == nil || len(snap.Tables) == 0 || len(other.Tables) == 0 {
		printer("(at least) one of the arguments is nil\n")
		return
	}

	for name, snapDesc := range snap.Tables {
		if otherDescr, present := other.Tables[name]; present {
			differenciate(name, snapDesc.Data, otherDescr.Data, printer.WithHeader("Table: [%s]\n", name))
		} else {
			printer("Table [%s] exists only in left snapshot (it was deleted)\n", name)
		}
	}

	for name, _ := range other.Tables {
		if _, present := snap.Tables[name]; !present {
			printer("Table [%s] exists only in right snapshot (it was added)\n", name)
		}
	}
}

func differenciate(name string, left, right msm2s, printer prn.Printer) {
	if len(left) != len(right) {
		printer("left and Right are different in size\n")
	}
	for key, lvalue := range left {
		if rvalue, exists := right[key]; exists {
			compare(key, lvalue, rvalue, printer.WithHeader("\tkey [%s]\n", key))
		} else {
			printer("\tkey [%s] only present on the left\n", key)
		}
	}
	for key, _ := range right {
		if _, exists := left[key]; !exists {
			printer("\tkey [%s] only present on the right\n", key)
		}
	}
}

func compare(element string, left, right m2s, printer prn.Printer) {
	for key, lvalue := range left {
		if rvalue, exists := right[key]; exists {
			if lvalue != rvalue {
				printer("\t\tfield: [%s]: [%s] ==>> [%s]\n", key, lvalue, rvalue)
			}
		} else {
			printer("\t\tfield: [%s] only present on the left\n", key)
		}
	}
	for key, _ := range right {
		if _, exists := left[key]; !exists {
			printer("\t\tfield: [%s] only present on the right\n", key)
		}
	}
}

func fillinSubstitutes() {

	actual := make(map[string]string)
	for key, _ := range substitutes {
		resolvesTo := os.Getenv(key)
		if len(resolvesTo) > 0 {
			actual[key] = resolvesTo
		}
	}
	substitutes = actual
}

func normalizeString(src string) string {
	if len(src) > 0 && len(substitutes) > 0 {
		for key, value := range substitutes {
			src = strings.ReplaceAll(src, value, "${"+key+"}")
		}
	}
	return src
}

func lookat(svc *dynamodb.DynamoDB, table string, snapshot *Snapshot) error {
	input := dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	}
	result, err := svc.DescribeTable(&input)
	if err != nil {
		return err
	}

	fmt.Println("------------------------------------- processing", *result.Table.TableName, "/", *result.Table.ItemCount)
	if *result.Table.ItemCount > maxAllowedItemCount {
		fmt.Println("table", *result.Table.TableName, "is too big --> ignoring it: ", *result.Table.ItemCount)
		// return fmt.Errorf("Table (%s) is too big (%v)", *result.Table.TableName, *result.Table.ItemCount)
		return nil // skipping a big table is not an error
	}

	desc := TableDesc{
		Name:           normalizeString(*result.Table.TableName),
		TableArn:       normalizeString(*result.Table.TableArn),
		KeySchema:      result.Table.KeySchema,
		TableSizeBytes: *result.Table.TableSizeBytes,
		ItemCount:      *result.Table.ItemCount,
	}

	primary := make([]string, 0, len(result.Table.KeySchema))
	for _, one := range result.Table.KeySchema {
		primary = append(primary, *one.AttributeName)
	}
	err = look_inside(svc, table, primary, &desc)
	if err != nil {
		return err
	}

	snapshot.Tables[normalizeString(table)] = desc
	return err
}

func look_inside(svc *dynamodb.DynamoDB, table string, keys []string, desc *TableDesc) error {
	input := &dynamodb.ScanInput{
		TableName: aws.String(table),
	}

	result, err := svc.Scan(input)
	if err != nil {
		return err
	}

	// fmt.Println(result)

	storage := make(msm2s)
	for _, i := range result.Items {
		item := m2s{}

		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return err
		} else {
			// fmt.Println(item)
			storage[get_key(item, keys)] = item
		}
	}
	desc.Data = storage
	return err
}

func get_key(payload m2s, keys []string) string {
	values := make([]string, len(keys), len(keys))
	for i, key := range keys {
		values[i] = payload[key]
	}
	return strings.Join(values, ":")
}

func init() {
	fillinSubstitutes()
}

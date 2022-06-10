package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string
type users []user
type user struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var (
	operationError = errors.New("-operation flag has to be specified")
	filenameError  = errors.New("-fileName flag has to be specified")
	itemError      = errors.New("-item flag has to be specified")
	idError        = errors.New("-id flag has to be specified")
)

func Perform(args Arguments, writer io.Writer) (err error) {
	if v, ok := args["operation"]; v == "" || !ok {
		return operationError
	}
	if v, ok := args["fileName"]; v == "" || !ok {
		return filenameError
	}

	switch args["operation"] {
	case "list":
		err = list(args, writer)
	case "add":
		if v, ok := args["item"]; v == "" || !ok {
			return itemError
		}
		err = add(args)
	case "remove":
		if v, ok := args["id"]; v == "" || !ok {
			return idError
		}
		err = remove(args)
	case "findById":
		if v, ok := args["id"]; v == "" || !ok {
			return idError
		}
		err = findById(args, writer)
	default:
		return fmt.Errorf("Operation %v not allowed!", args["operation"])
	}

	if err != nil {
		_, err = writer.Write([]byte(err.Error()))
	}

	return
}

func list(args Arguments, writer io.Writer) error {
	us, _, err := readAll(args["fileName"])
	if err != nil {
		return err
	}

	b, err := json.Marshal(us)
	if err != nil {
		return err
	}

	if len(us) <= 0 {
		b = []byte{}
	}

	_, err = writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func add(args Arguments) error {
	us, f, err := readAll(args["fileName"])
	if err != nil {
		return err
	}

	var u user
	err = json.Unmarshal([]byte(args["item"]), &u)
	if err != nil {
		return err
	}

	for _, u2 := range us {
		if u2.Id == u.Id {
			return fmt.Errorf("Item with id %v already exists", u2.Id)
		}
	}

	us = append(us, u)

	err = writeToFile(us, f)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func remove(args Arguments) error {
	us, f, err := readAll(args["fileName"])
	if err != nil {
		return err
	}

	var found bool
	for k, u := range us {
		if u.Id == args["id"] {
			us = append(us[:k], us[k+1:]...)
			found = true
		}
	}

	if !found {
		return fmt.Errorf("Item with id %v not found", args["id"])
	}

	err = writeToFile(us, f)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func findById(args Arguments, writer io.Writer) error {
	us, _, err := readAll(args["fileName"])
	if err != nil {
		return err
	}

	var found bool
	var b []byte
	for _, u := range us {
		if u.Id == args["id"] {
			b, err = json.Marshal(u)
			if err != nil {
				return err
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Item with id %v not found", args["id"])
	}

	_, err = writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func readAll(fileName string) (us users, f *os.File, err error) {
	f, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &us)
		if err != nil {
			return nil, nil, err
		}
	}

	return
}

func writeToFile(us users, f *os.File) (err error) {
	b, err := json.Marshal(us)
	if err != nil {
		return err
	}

	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.Write(b)

	return
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func parseArgs() (args Arguments) {
	args["id"] = *flag.String("id", "", "item flag")
	args["item"] = *flag.String("item", "", "item flag")
	args["operation"] = *flag.String("operation", "", "operation flag")
	args["fileName"] = *flag.String("fileName", "", "fileName flag")
	flag.Parse()
	return
}

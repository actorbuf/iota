package excel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type ImportReturn struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

const ImportSuccessCode = 1

func ToPhpExcel(ctx context.Context, file io.Reader) ([][]string, error) {
	url := "https://robotcrmapi.teammvp.art/server/excel"
	method := "POST"
	var returnData [][]string

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part1, err := writer.CreateFormFile("excel_data", "excel_data_file.xlsx")
	if err != nil {
		return returnData, err
	}
	_, err = io.Copy(part1, file)
	if err != nil {
		return returnData, err
	}
	err = writer.Close()
	if err != nil {
		return returnData, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return returnData, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return returnData, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != 200 {
		return returnData, errors.New("request excel serve error")
	}

	var importReturn ImportReturn
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return returnData, err
	}
	err = json.Unmarshal(body, &importReturn)
	if err != nil {
		return returnData, err
	}
	if importReturn.Code != ImportSuccessCode {
		return returnData, errors.New("import err " + importReturn.Msg)
	}
	returnData = importReturn.Data

	return returnData, nil
}

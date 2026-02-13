package groupmeext

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/beeper/groupme-lib"
	log "maunium.net/go/maulogger/v2"
)

type Message struct{ groupme.Message }

func (m *Message) Scan(value interface{}) error {
	bytes, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal json value:", value))
	}

	message := Message{}
	err := json.Unmarshal([]byte(bytes), &message)

	*m = Message(message)
	return err
}

func (m *Message) Value() (driver.Value, error) {
	e, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// DownloadImage helper function to download image from groupme;
// append .large/.preview/.avatar to get various sizes
func DownloadImage(URL string, log log.Logger) (bytes *[]byte, mime string, err error) {
	//TODO check its actually groupme?
	response, err := http.Get(URL)
	if err != nil {
		log.Errorln("Failed to download image:", err)
		return nil, "", errors.New("Failed to download avatar: " + err.Error())
	}
	defer response.Body.Close()

	image, err := ioutil.ReadAll(response.Body)
	bytes = &image
	if err != nil {
		log.Errorln("Failed to read image body:", err)
		return nil, "", errors.New("Failed to read downloaded image:" + err.Error())
	}

	mime = response.Header.Get("Content-Type")
	if len(mime) == 0 {
		mime = http.DetectContentType(image)
	}
	return
}

func DownloadFile(RoomJID groupme.ID, FileID string, token string, log log.Logger) (contents []byte, fname, mime string, err error) {
	client := &http.Client{}
	b, _ := json.Marshal(struct {
		FileIDS []string `json:"file_ids"`
	}{
		FileIDS: []string{FileID},
	})

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://file.groupme.com/v1/%s/fileData", RoomJID), bytes.NewReader(b))
	req.Header.Add("X-Access-Token", token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Failed to get file data:", err)
		return nil, "", "", err
	}

	defer resp.Body.Close()
	data := []ImgData{}
	json.NewDecoder(resp.Body).Decode(&data)
	// fmt.Println(data, RoomJID, FileID, token)
	if len(data) < 1 {
		log.Warnln("No file data found for", FileID)
		return nil, "", "", errors.New("no file data found")
	}

	req, _ = http.NewRequest("POST", fmt.Sprintf("https://file.groupme.com/v1/%s/files/%s", RoomJID, FileID), nil)
	req.URL.Query().Add("token", token)
	req.Header.Add("X-Access-Token", token)
	resp, err = client.Do(req)
	if err != nil {
		log.Errorln("Failed to download file:", err)
		return nil, "", "", err
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)
	return bytes, data[0].FileData.FileName, data[0].FileData.Mime, nil
}

func DownloadVideo(previewURL, videoURL, token string, log log.Logger) (vidContents []byte, mime string, err error) {
	//preview TODO
	client := &http.Client{}

	req, _ := http.NewRequest("GET", videoURL, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Failed to download video:", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("failed to download video: status %d", resp.StatusCode)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	mime = resp.Header.Get("Content-Type")
	if len(mime) == 0 {
		mime = http.DetectContentType(bytes)
	}
	return bytes, mime, nil

}

type ImgData struct {
	FileData struct {
		FileName string `json:"file_name"`
		FileSize int    `json:"file_size"`
		Mime     string `json:"mime_type"`
	} `json:"file_data"`
}

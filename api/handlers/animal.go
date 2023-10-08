package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	animal "anidex_api/domain/animal"
)

type AnimalRequest struct {
	Photos      string `json:"photo"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Taxonomy    string `json:"taxonomy"`
	Etymology   string `json:"etymology"`
	Iucn        string `json:"iucn"`
	Geo         string `json:"geo"`
	Migration   string `json:"migration"`
	Habitat     string `json:"habitat"`
	Dimensions  string `json:"dimensions"`
	Ds          string `json:"ds"`
	Diet        string `json:"diet"`
	Description string `json:"description"`
}

type AnimalResponse struct {
	AnimalData AnimalRequest `json:"animalData"`
	Error      string        `json:"error"`
	Message    string        `json:"message"`
	Status     int           `json:"status"`
}

func getDataFromMap(key string, originData map[string][]string) ([]string, error) {
	if data, exists := originData[key]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("%s is not present in provided data", key)
}

func (ar *AnimalRequest) readAnimalRequestValues(values map[string][]string) error {
	for _, v := range animal.ANIMAL_POST_REQUEST_MANDATORY_FIELDS {
		if v == "photo[]" {
			continue
		}

		var structFieldName string
		if v == "iucn[]" {
			structFieldName = strings.ToUpper(v[0:1]) + v[1:]
			structFieldName = strings.ReplaceAll(structFieldName, "[]", "")
		} else {
			structFieldName = strings.ToUpper(v[0:1]) + v[1:]
		}

		value, err := getDataFromMap(v, values)

		if v == "description" && err != nil {
			ar.Description = ""
			continue
		}

		if v == "iucn[]" && err == nil {
			ar.Iucn = strings.Join(value, ",")
			continue
		}

		if err != nil {
			return err
		}

		reflect.ValueOf(ar).Elem().FieldByName(structFieldName).SetString(value[0])
	}

	return nil
}

func animalRequestErrorResponse(w http.ResponseWriter, err error) []byte {
	response := AnimalResponse{
		Error:   err.Error(),
		Status:  http.StatusBadRequest,
		Message: "",
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return jsonResponse
}

func animalRequestSuccessResponse(w http.ResponseWriter, a AnimalRequest) []byte {
	response := AnimalResponse{
		AnimalData: a,
		Status:     http.StatusCreated,
		Message:    "New animal species successfully registered in the Anidex",
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return jsonResponse
}

func storePhotosAndCollectPaths(m *multipart.Form) ([]string, int32, []error) {
	var photoPaths []string
	var storedPhotosCount int32
	var errs []error

	for _, v := range m.File["photo[]"] {
		file, err := v.Open()
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()

		// Create directory
		//dirPath := "/Users/accilo/Desktop/angelo/anidex_api/temp-img"
		dirPath := os.Getenv("RAILWAY_VOLUME_MOUNT_PATH") + "/uploaded_images"
		if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(dirPath, os.ModePerm)
			if err != nil {
				fmt.Println(err)
				errs = append(errs, err)
				continue
			}
		}

		newFilePath := filepath.Join(dirPath, filepath.Base(v.Filename))

		// Create file
		dst, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
			errs = append(errs, err)
			continue
		}
		defer dst.Close()

		if _, err = io.Copy(dst, file); err != nil {
			fmt.Println(err)
			errs = append(errs, err)
			continue
		}

		photoPaths = append(photoPaths, newFilePath)
		storedPhotosCount += 1
	}
	fmt.Println("Successfully Uploaded Image")
	return photoPaths, storedPhotosCount, errs
}

func (ar *AnimalRequest) buildAnimalRequest(m *multipart.Form) error {

	photoPaths, uploadedPhotosCount, errs := storePhotosAndCollectPaths(m)

	if len(errs) > 0 {
		return fmt.Errorf("The upload of photos produced an error: %v", errs)
	}

	if len(photoPaths) != int(uploadedPhotosCount) {
		return fmt.Errorf("The upload of photos produced an error: %v", errs)
	}

	ar.Photos = strings.Join(photoPaths, ",")

	err := ar.readAnimalRequestValues(m.Value)
	if err != nil {
		return err
	}

	return nil

}

func PostAnimal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Content-Type", "application/json")
	r.ParseMultipartForm(10 << 20)

	var animalRequest AnimalRequest
	err := animalRequest.buildAnimalRequest(r.MultipartForm)
	if err != nil {
		w.Write(animalRequestErrorResponse(w, err))
		return
	}

	//retrieve DB from context
	db := r.Context().Value("db").(*sql.DB)

	//save data in mysql
	stmt, err := db.Prepare("INSERT INTO animals (photos,name,taxonomy,etymology,iucn,geo,migration,habitat,dimensions,ds,diet,description,category) VALUES (?, ?,?, ?,?, ?,?, ?,?, ?,?, ?,?)")
	if err != nil {
		w.Write(animalRequestErrorResponse(w, err))
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		animalRequest.Photos,
		animalRequest.Name,
		animalRequest.Taxonomy,
		animalRequest.Etymology,
		animalRequest.Iucn,
		animalRequest.Geo,
		animalRequest.Migration,
		animalRequest.Habitat,
		animalRequest.Dimensions,
		animalRequest.Ds,
		animalRequest.Diet,
		animalRequest.Description,
		animalRequest.Category,
	)

	if err != nil {
		w.Write(animalRequestErrorResponse(w, err))
		return
	}

	w.Write(animalRequestSuccessResponse(w, animalRequest))

}
package groupietracker

type Data struct {
	Id int `json :"id"`
	DatesLocation map[string][]string `json :"datesLocation"`
}

func web_app() {

}
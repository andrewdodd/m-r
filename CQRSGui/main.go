package main

import (
	s "SimpleCQRS/SimpleCQRS"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"reflect"
	"strconv"
)

func addTemplates(templates map[string]*template.Template) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "templates", templates)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getTemplates(r *http.Request) map[string]*template.Template {
	return r.Context().Value("templates").(map[string]*template.Template)
}

func addReadModel(rm s.ReadModel) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "readmodel", rm)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getReadModel(r *http.Request) s.ReadModel {
	return r.Context().Value("readmodel").(s.ReadModel)
}

func addBus(bus *s.FakeBus) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "bus", bus)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getBus(r *http.Request) *s.FakeBus {
	return r.Context().Value("bus").(*s.FakeBus)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	template := getTemplates(r)["index"]
	readmodel := getReadModel(r)
	data := make(map[string][]s.InventoryItemListDto)
	data["InventoryItems"] = readmodel.GetInventoryItems()
	err := template.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		template := getTemplates(r)["add"]
		err := template.ExecuteTemplate(w, "base", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		name := r.FormValue("name")
		bus := getBus(r)
		err := bus.Dispatch(s.CreateInventoryItem{InventoryItemId: s.NewGuid(), Name: name})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func changeNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	readmodel := getReadModel(r)
	ii, err := readmodel.GetInventoryItemDetails(s.Guid(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		template := getTemplates(r)["changename"]
		err := template.ExecuteTemplate(w, "base", ii)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		name := r.FormValue("name")
		v, err := strconv.ParseInt(r.FormValue("version"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		version := int(v)
		bus := getBus(r)
		err = bus.Dispatch(s.RenameInventoryItem{ii.Id, version, name})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func checkinHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	readmodel := getReadModel(r)
	ii, err := readmodel.GetInventoryItemDetails(s.Guid(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		template := getTemplates(r)["checkin"]
		err := template.ExecuteTemplate(w, "base", ii)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		number, err := strconv.Atoi(r.FormValue("number"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		version, err := strconv.Atoi(r.FormValue("version"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bus := getBus(r)
		err = bus.Dispatch(s.CheckInItemsToInventory{ii.Id, version, number})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func removeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	readmodel := getReadModel(r)
	ii, err := readmodel.GetInventoryItemDetails(s.Guid(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		template := getTemplates(r)["remove"]
		err := template.ExecuteTemplate(w, "base", ii)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		number, err := strconv.Atoi(r.FormValue("number"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		version, err := strconv.Atoi(r.FormValue("version"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bus := getBus(r)
		err = bus.Dispatch(s.RemoveItemsFromInventory{ii.Id, version, number})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func deactivateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	readmodel := getReadModel(r)
	ii, err := readmodel.GetInventoryItemDetails(s.Guid(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		template := getTemplates(r)["deactivate"]
		err := template.ExecuteTemplate(w, "base", ii)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		id = r.FormValue("id")
		version, err := strconv.Atoi(r.FormValue("version"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bus := getBus(r)
		err = bus.Dispatch(s.DeactivateInventoryItem{s.Guid(id), version})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	template := getTemplates(r)["details"]
	readmodel := getReadModel(r)
	vars := mux.Vars(r)
	id := vars["id"]

	guid := s.Guid(id)
	data := make(map[string]s.InventoryItemDetailsDto)
	model, _ := readmodel.GetInventoryItemDetails(guid)
	data["Model"] = model
	err := template.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setupCQRS(mimicEventualConsistency bool) (s.ReadModel, *s.FakeBus) {

	bus := s.NewFakeBus(mimicEventualConsistency)
	storage := s.NewEventStore(&bus)
	rep := s.InventoryItemRepository{storage}
	commands := s.NewInventoryCommandHandlers(rep)
	bus.SetCommandHandler(reflect.TypeOf(s.CheckInItemsToInventory{}), commands.HandleCheckInItemsToInventory)
	bus.SetCommandHandler(reflect.TypeOf(s.CreateInventoryItem{}), commands.HandleCreateInventoryItem)
	bus.SetCommandHandler(reflect.TypeOf(s.DeactivateInventoryItem{}), commands.HandleDeactivateInventoryItem)
	bus.SetCommandHandler(reflect.TypeOf(s.RemoveItemsFromInventory{}), commands.HandleRemoveItemsFromInventory)
	bus.SetCommandHandler(reflect.TypeOf(s.RenameInventoryItem{}), commands.HandleRenameInventoryItem)

	bsdb := s.NewBSDB()

	detail := s.NewInventoryItemDetailView(&bsdb)
	bus.AddEventProcessor(reflect.TypeOf(s.InventoryItemCreated{}), detail.ProcessInventoryItemCreated)
	bus.AddEventProcessor(reflect.TypeOf(s.InventoryItemDeactivated{}), detail.ProcessInventoryItemDeactivated)
	bus.AddEventProcessor(reflect.TypeOf(s.InventoryItemRenamed{}), detail.ProcessInventoryItemRenamed)
	bus.AddEventProcessor(reflect.TypeOf(s.ItemsCheckedInToInventory{}), detail.ProcessItemsCheckedInToInventory)
	bus.AddEventProcessor(reflect.TypeOf(s.ItemsRemovedFromInventory{}), detail.ProcessItemsRemovedFromInventory)

	list := s.NewInventoryListView(&bsdb)
	bus.AddEventProcessor(reflect.TypeOf(s.InventoryItemCreated{}), list.ProcessInventoryItemCreated)
	bus.AddEventProcessor(reflect.TypeOf(s.InventoryItemRenamed{}), list.ProcessInventoryItemRenamed)
	bus.AddEventProcessor(reflect.TypeOf(s.InventoryItemDeactivated{}), list.ProcessInventoryItemDeactivated)

	id := s.NewGuid()
	bus.Dispatch(s.CreateInventoryItem{id, "The self-seed inventory item"})
	rmf := s.NewReadModelFacade(&bsdb)
	return &rmf, &bus
}

func buildTemplates() map[string]*template.Template {
	t := make(map[string]*template.Template)

	for _, name := range []string{"index", "details", "add", "changename", "checkin", "remove", "deactivate"} {
		t[name] = template.Must(
			template.ParseFiles(
				fmt.Sprintf("./CQRSGui/pages/%v.html", name),
				"./CQRSGui/pages/base.html"))
	}

	return t
}

func main() {
	templates := buildTemplates()

	readmodel, bus := setupCQRS(false) // true to introduce delays
	rtr := mux.NewRouter()
	rtr.Use(addReadModel(readmodel))
	rtr.Use(addTemplates(templates))
	rtr.Use(addBus(bus))

	rtr.HandleFunc("/", indexHandler).Methods("GET")
	rtr.HandleFunc("/add", addHandler).Methods("GET", "POST")

	ii := rtr.PathPrefix("/details").Subrouter()
	ii.HandleFunc("/{id}", detailsHandler).Methods("GET")
	ii.HandleFunc("/{id}/changename", changeNameHandler).Methods("GET", "POST")
	ii.HandleFunc("/{id}/checkin", checkinHandler).Methods("GET", "POST")
	ii.HandleFunc("/{id}/remove", removeHandler).Methods("GET", "POST")
	ii.HandleFunc("/{id}/deactivate", deactivateHandler).Methods("GET", "POST")

	rtr.PathPrefix("/Content/").Handler(
		http.StripPrefix("/Content/",
			http.FileServer(http.Dir("./CQRSGui/Content/"))))

	http.Handle("/", rtr)
	http.ListenAndServe(":8080", nil)
}

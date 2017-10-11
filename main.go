package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type prov struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type reg struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type dis struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type vil struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	ZipCode string `json:"zip_code"`
}

// Province consists of list of regencies
type Province struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Regencies []Regency `json:"regencies"`
}

// Regency consists of list of districts
type Regency struct {
	ID         int        `json:"id"`
	ProvinceID int        `json:"province_id"`
	Name       string     `json:"name"`
	Districts  []District `json:"districts"`
}

// District consists of list of villages
type District struct {
	ID         int       `json:"id"`
	RegencyID  int       `json:"regency_id"`
	ProvinceID int       `json:"province_id"`
	Name       string    `json:"name"`
	Villages   []Village `json:"villages"`
}

// Village consists of name and zip
type Village struct {
	ID         int    `json:"id"`
	DistrictID int    `json:"district_id"`
	RegencyID  int    `json:"regency_id"`
	ProvinceID int    `json:"province_id"`
	Name       string `json:"name"`
	ZipCode    string `json:"zip_code"`
}

// Change if the number of max province is changed
const maxProv = 34
const provDataLength = 10
const provProvIdx = 1

// Change if the number of max regency per province is changed
const maxRegPerProv = 9
const regDataLength = 7
const regProvIdx = 1
const regRegPrefixIdx = 2
const regRegIdx = 3

// Change if the number of max village is changed
const maxVil = 82505
const vilDataLength = 6
const vilZipCodeIdx = 1
const vilVilIdx = 2
const vilDisIdx = 3
const vilRegPrefixIdx = 4
const vilRegIdx = 5

var wg sync.WaitGroup

func main() {
	// Comment this if you already scrap province
	// scrapProvince(100, 5)

	fmt.Println("Populate Province.")
	listOfProvStr := populateProvince(100)
	fmt.Printf("Finish Populate Province. Len() = %v.\n", len(listOfProvStr))

	// Comment this if you already scrap regency
	// scrapRegency(100, 5, listOfProvStr)

	fmt.Println("Populate Regency.")
	regProvMap := populateRegency(100, listOfProvStr)
	fmt.Printf("Finish Populate Regency. Len() = %v.\n", len(regProvMap))

	// Comment this if you already scrap village
	// scrapVillage(100, 5)

	fmt.Println("Populate Village.")
	zipCodes, listOfProv, regMap, disMap, vilMap, zipCodeMap := populateVillage(100, regProvMap)
	fmt.Println("Finish Populate Village")

	fmt.Println("Parse And Save JSON")
	parseAndSaveJSON(zipCodes, listOfProv, regMap, disMap, vilMap, zipCodeMap)
}

func scrapProvince(numOfData int, numOfConc int) {
	for i, j := 0, 0; i <= maxProv/numOfData; i++ {
		if j == numOfConc {
			j = 0
			time.Sleep(5 * time.Second)
		}
		j++

		wg.Add(1)
		go func(n int, numOfData int) {
			if maxProv == numOfData*n {
				return
			}

			file, err := os.Create(fmt.Sprintf("file/list_province_%v.txt", n))
			if err != nil {
				panic(err)
			}
			defer file.Close()

			doc, err := goquery.NewDocument(fmt.Sprintf("http://kodepos.nomor.net/_kodepos.php?_i=provinsi-kodepos&daerah=&jobs=&perhal=%v&sby=000000&no1=%v&no2=%v", numOfData, (numOfData*n)-(numOfData-1), numOfData*n))
			if err != nil {
				panic(err)
			}

			fmt.Printf("Processing %v - %v\n", n*numOfData, (n+1)*numOfData)

			doc.Find("html > body > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr > td > table").Eq(2).Find("tr[bgcolor='#ccffff'] > td").Each(func(n int, s *goquery.Selection) {
				fmt.Fprintln(file, s.Text())
			})

			wg.Done()
		}(i, numOfData)
	}
	wg.Wait()
}

func populateProvince(numOfData int) []string {
	var listOfProvStr []string
	for i := 0; i <= maxProv/numOfData; i++ {
		body, err := ioutil.ReadFile(fmt.Sprintf("file/list_province_%v.txt", i))
		if err != nil {
			panic(err)
		}

		rows := strings.Split(string(body), "\n")

		for i = 0; i < len(rows)/provDataLength; i++ {
			provStr := rows[(i*provDataLength)+provProvIdx]
			listOfProvStr = append(listOfProvStr, provStr)
		}
	}
	return listOfProvStr
}

func scrapRegency(numOfData int, numOfConc int, listOfProvStr []string) {
	for x := 0; x < len(listOfProvStr); x++ {
		for i, j := 0, 0; i <= maxRegPerProv/numOfData; i++ {
			if j == numOfConc {
				j = 0
				time.Sleep(5 * time.Second)
			}
			j++

			wg.Add(1)
			go func(n int, numOfData int, prov string) {
				if maxRegPerProv == numOfData*n {
					return
				}

				file, err := os.Create(fmt.Sprintf("file/province_%v_regency_%v.txt", prov, n))
				if err != nil {
					panic(err)
				}
				defer file.Close()

				doc, err := goquery.NewDocument(fmt.Sprintf("http://kodepos.nomor.net/_kodepos.php?_i=kota-kodepos&daerah=Provinsi&perhal=%v&sby=000000&no1=%v&no2=%v&jobs=%v", numOfData, (numOfData*n)-(numOfData-1), numOfData*n, url.PathEscape(prov)))
				if err != nil {
					panic(err)
				}

				fmt.Printf("Processing %v - %v\n", n*numOfData, (n+1)*numOfData)

				doc.Find("html > body > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr > td > table").Eq(2).Find("tr[bgcolor='#ccffff'] > td").Each(func(n int, s *goquery.Selection) {
					fmt.Fprintln(file, s.Text())
				})

				wg.Done()
			}(i, numOfData, listOfProvStr[x])
		}
	}
	wg.Wait()
}

func populateRegency(numOfData int, listOfProvStr []string) map[string]Province {
	regProvMap := make(map[string]Province)
	for x := 0; x < len(listOfProvStr); x++ {
		for i := 0; i <= maxRegPerProv/numOfData; i++ {
			body, err := ioutil.ReadFile(fmt.Sprintf("file/province_%v_regency_%v.txt", listOfProvStr[x], i))
			if err != nil {
				panic(err)
			}

			rows := strings.Split(string(body), "\n")

			for i = 0; i < len(rows)/regDataLength; i++ {
				provStr := rows[(i*regDataLength)+regProvIdx]
				regPrefix := rows[(i*regDataLength)+regRegPrefixIdx]
				regStr := rows[(i*regDataLength)+regRegIdx]

				regProvMap[regPrefix+" "+regStr] = Province{0, provStr, nil}
			}
		}
	}
	return regProvMap
}

func scrapVillage(numOfData int, numOfConc int) {
	for i, j := 0, 0; i <= maxVil/numOfData; i++ {
		if j == numOfConc {
			j = 0
			time.Sleep(5 * time.Second)
		}
		j++

		wg.Add(1)
		go func(n int, numOfData int) {
			if maxVil == numOfData*n {
				return
			}

			file, err := os.Create(fmt.Sprintf("file/village_%v.txt", n))
			if err != nil {
				panic(err)
			}
			defer file.Close()

			doc, err := goquery.NewDocument(fmt.Sprintf("http://kodepos.nomor.net/_kodepos.php?_i=desa-kodepos&perhal=%v&sby=000000&no1=%v&no2=%v", numOfData, (numOfData*n)-(numOfData-1), numOfData*n))
			if err != nil {
				panic(err)
			}

			fmt.Printf("Processing %v - %v\n", n*numOfData, (n+1)*numOfData)

			doc.Find("html > body > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr > td > table").Eq(2).Find("tr[bgcolor='#ccffff'] > td").Each(func(n int, s *goquery.Selection) {
				fmt.Fprintln(file, s.Text())
			})

			wg.Done()
		}(i, numOfData)
	}
	wg.Wait()
}

func populateVillage(numOfData int, regProvMap map[string]Province) ([]Province, []prov, map[string][]reg, map[string][]dis, map[string][]vil, map[string]string) {
	m := make(map[string]map[string]map[string]map[string]string)
	var listOfProv []prov
	mReg := make(map[string][]reg)
	mDis := make(map[string][]dis)
	mVil := make(map[string][]vil)
	mZipCode := make(map[string]string)

	for i := 0; i <= maxVil/numOfData; i++ {
		body, err := ioutil.ReadFile(fmt.Sprintf("file/village_%v.txt", i))
		if err != nil {
			panic(err)
		}

		rows := strings.Split(string(body), "\n")

		for j := 0; j < len(rows)/vilDataLength; j++ {
			zipCode := strings.Split(rows[(j*vilDataLength)+vilZipCodeIdx], " ")[2]
			vilStr := rows[(j*vilDataLength)+vilVilIdx]
			disStr := rows[(j*vilDataLength)+vilDisIdx]
			regPrefix := rows[(j*vilDataLength)+vilRegPrefixIdx]
			regStr := rows[(j*vilDataLength)+vilRegIdx]
			prov := regProvMap[regPrefix+" "+regStr]

			// Hotfix
			if disStr == "Kinovaru" {
				disStr = "Kinovaro"
			}

			if _, ok := m[prov.Name]; !ok {
				m[prov.Name] = make(map[string]map[string]map[string]string)
			}

			if _, ok := m[prov.Name][regPrefix+" "+regStr]; !ok {
				m[prov.Name][regPrefix+" "+regStr] = make(map[string]map[string]string)
			}

			if _, ok := m[prov.Name][regPrefix+" "+regStr][disStr]; !ok {
				m[prov.Name][regPrefix+" "+regStr][disStr] = make(map[string]string)
			}

			if _, ok := m[prov.Name][regPrefix+" "+regStr][disStr][vilStr]; !ok {
				m[prov.Name][regPrefix+" "+regStr][disStr][vilStr] = zipCode
			}
		}
	}

	var zipCodes []Province
	provID, regID, disID, vilID := 1, 1, 1, 1

	for provName, mmProv := range m {
		var regs []Regency
		for regName, mmReg := range mmProv {
			var diss []District
			for disName, mmDis := range mmReg {
				var vils []Village
				for vilName, zipCode := range mmDis {
					vils = append(vils, Village{vilID, disID, regID, provID, vilName, zipCode})
					mVil[strconv.Itoa(disID)] = append(mVil[strconv.Itoa(disID)], vil{vilID, vilName, zipCode})
					mZipCode[strconv.Itoa(vilID)] = zipCode

					vilID++
				}
				diss = append(diss, District{disID, regID, provID, disName, vils})
				mDis[strconv.Itoa(regID)] = append(mDis[strconv.Itoa(regID)], dis{disID, disName})

				disID++
			}
			regs = append(regs, Regency{regID, provID, regName, diss})
			mReg[strconv.Itoa(provID)] = append(mReg[strconv.Itoa(provID)], reg{regID, regName})

			regID++
		}
		zipCodes = append(zipCodes, Province{provID, provName, regs})
		listOfProv = append(listOfProv, prov{provID, provName})

		provID++
	}

	fmt.Println(provID, regID, disID, vilID)

	return zipCodes, listOfProv, mReg, mDis, mVil, mZipCode
}

func parseAndSaveJSON(zipCodes []Province, listOfProv []prov, regMap map[string][]reg, disMap map[string][]dis, vilMap map[string][]vil, zipCodeMap map[string]string) {
	file, err := os.Create("zip_codes.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	err = enc.Encode(&zipCodes)
	if err != nil {
		panic(err)
	}

	file2, err := os.Create("prov_map.txt")
	if err != nil {
		panic(err)
	}
	defer file2.Close()

	enc2 := json.NewEncoder(file2)
	err = enc2.Encode(&listOfProv)
	if err != nil {
		panic(err)
	}

	file3, err := os.Create("reg_map.txt")
	if err != nil {
		panic(err)
	}
	defer file3.Close()

	enc3 := json.NewEncoder(file3)
	err = enc3.Encode(&regMap)
	if err != nil {
		panic(err)
	}

	file4, err := os.Create("dis_map.txt")
	if err != nil {
		panic(err)
	}
	defer file4.Close()

	enc4 := json.NewEncoder(file4)
	err = enc4.Encode(&disMap)
	if err != nil {
		panic(err)
	}

	file5, err := os.Create("vil_map.txt")
	if err != nil {
		panic(err)
	}
	defer file5.Close()

	enc5 := json.NewEncoder(file5)
	err = enc5.Encode(&vilMap)
	if err != nil {
		panic(err)
	}

	file6, err := os.Create("zip_code_map.txt")
	if err != nil {
		panic(err)
	}
	defer file6.Close()

	enc6 := json.NewEncoder(file6)
	err = enc6.Encode(&zipCodeMap)
	if err != nil {
		panic(err)
	}
}

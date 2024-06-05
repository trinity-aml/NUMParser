package main

import (
	"NUMParser/config"
	"NUMParser/db"
	"NUMParser/db/models"
	"NUMParser/movies/tmdb"
	"NUMParser/parser"
	"NUMParser/releases"
	"NUMParser/version"
	"NUMParser/web"
	"context"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/jasonlvhit/gocron"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type args struct {
	Port     string `arg:"-p" help:"web server port, default 38888"`
	Proxy    string `arg:"--proxy" help:"proxy for rutor, http://user:password@ip:port"`
	UseProxy bool   `arg:"--useproxy" help:"enable auto proxy"`
}

var params args

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	arg.MustParse(&params)

	fmt.Println("=========== START ===========")
	fmt.Println("numParser_"+version.Version+",", runtime.Version()+",", "CPU Num:", runtime.NumCPU())

	if params.Port == "" {
		port, err := config.ReadConfigParser("Port")
		if err == nil {
			params.Port = port
		} else {
			params.Port = "38888"
		}
	}

	if params.Proxy != "" {
		_, err := url.Parse(params.Proxy)
		if err != nil {
			log.Println("Error parse proxy host:", err)
		} else {
			config.ProxyHost = params.Proxy
		}
	} else {
		proxy, err := config.ReadConfigParser("Proxy")
		if err == nil {
			params.Proxy = proxy
			_, err := url.Parse(params.Proxy)
			if err == nil {
				config.ProxyHost = params.Proxy
			} else {
				log.Println("Error parse proxy host:", err)
			}
		}
	}

	if params.UseProxy == true {
		config.UseProxy = params.UseProxy
	} else {
		use_proxy, err := config.ReadConfigParser("UseProxy")
		if err == nil && use_proxy == "true" {
			params.UseProxy = true
		} else {
			params.UseProxy = false
		}
	}

	dnsResolve()

	db.Init()
	loadProxy()
	tmdb.Init()

	getDbInfo()
	web.Start(params.Port)

	scanReleases()
	scanMoviesYears()
	web.SetStaticReleases()

	log.Println("Start timer")
	gocron.Every(3).Hours().From(calcTime()).Do(scanReleases)
	gocron.Every(1).Day().At("2:30").Do(scanMoviesYears)
	<-gocron.Start()

}

func dnsResolve() {
	hosts := [6]string{"1.1.1.1", "1.0.0.1", "208.67.222.222", "208.67.220.220", "8.8.8.8", "8.8.4.4"}
	ret := 0
	for _, ip := range hosts {
		ret = toolResolve("www.google.com", ip)
		switch {
		case ret == 2:
			fmt.Println("DNS resolver OK\n")
		case ret == 1:
			fmt.Println("New DNS resolver OK\n")
		case ret == 0:
			fmt.Println("New DNS resolver failed\n")
		}
		if ret == 2 || ret == 1 {
			break
		}
	}
}

func toolResolve(host string, serverDNS string) int {
	addrs, err := net.LookupHost(host)
	addr_dns := fmt.Sprintf("%s:53", serverDNS)
	a := 0
	if len(addrs) == 0 {
		fmt.Println("Check dns", addrs, err)
		fn := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", addr_dns)
		}
		net.DefaultResolver = &net.Resolver{
			Dial: fn,
		}
		addrs, err = net.LookupHost(host)
		fmt.Println("Check new dns", addrs, err)
		if err == nil || len(addrs) > 0 {
			a = 1
		} else {
			a = 0
		}
	} else {
		a = 2
	}
	return a
}

func scanReleases() {
	loadProxy()
	getDbInfo()
	rutorParser := parser.NewRutor()
	rutorParser.Parse()

	releases.GetNewMovies()
	releases.GetFourKMovies()
	releases.GetNewTVs()
	releases.GetNewCartoons()
	releases.GetNewCartoonsTV()
	db.SaveAll()
	copy()
}

func scanMoviesYears() {
	loadProxy()
	releases.GetLegends()
	for y := 1980; y <= time.Now().Year(); y++ {
		releases.GetNewMoviesYear(y)
	}
	db.SaveAll()
	copy()
}

// Exec script for copy any files
func copy() {
	dir := filepath.Dir(os.Args[0])
	_, err := os.Stat("copy.sh")
	if err == nil {
		logOut, err := exec.Command("/bin/sh", filepath.Join(dir, "copy.sh")).CombinedOutput()
		if err != nil {
			log.Println("Error copy releases:", err)
		}
		output := string(logOut)
		log.Println(output)
	}
}

// Exec script for load proxy, script mast create file proxy.list
func loadProxy() {
	if config.UseProxy {
		log.Println("Load proxy list...")
		dir := filepath.Dir(os.Args[0])
		logOut, err := exec.Command("/bin/sh", filepath.Join(dir, "proxy.sh")).CombinedOutput()
		if err != nil {
			log.Println("Error proxy releases:", err)
		}
		output := string(logOut)
		log.Println(output)
	}
}

func calcTime() *time.Time {
	//2 5 8 11 14 17 20 23
	hour := time.Now().Hour()
	t := time.Date(time.Now().Year(),
		time.Now().Month(),
		time.Now().Day(),
		0, 0, 0, 0, time.Local)
	if hour < 2 {
		t = t.Add(2 * time.Hour)
	} else if hour < 5 {
		t = t.Add(5 * time.Hour)
	} else if hour < 11 {
		t = t.Add(11 * time.Hour)
	} else if hour < 14 {
		t = t.Add(14 * time.Hour)
	} else if hour < 17 {
		t = t.Add(17 * time.Hour)
	} else if hour < 20 {
		t = t.Add(20 * time.Hour)
	} else if hour < 23 {
		t = t.Add(23 * time.Hour)
	} else if hour >= 23 {
		t = t.Add(26 * time.Hour)
	}
	return &t
}

func getDbInfo() {
	listTorr := db.GetTorrs()
	wIMDB := 0
	nIMDB := 0

	cMovie := 0
	cSeries := 0
	cDocMovie := 0
	cDocSeries := 0
	cCartoonMovie := 0
	cCartoonSeries := 0
	cTVShow := 0
	cAnime := 0

	for _, d := range listTorr {
		if d.IMDBID != "" {
			wIMDB++
		} else {
			nIMDB++
		}
		switch d.Categories {
		case models.CatMovie:
			cMovie++
		case models.CatSeries:
			cSeries++
		case models.CatDocMovie:
			cDocMovie++
		case models.CatDocSeries:
			cDocSeries++
		case models.CatCartoonMovie:
			cCartoonMovie++
		case models.CatCartoonSeries:
			cCartoonSeries++
		case models.CatTVShow:
			cTVShow++
		case models.CatAnime:
			cAnime++
		}
	}

	fmt.Println("Movies:", cMovie)
	fmt.Println("Serials:", cSeries)
	fmt.Println("Doc Movies:", cDocMovie)
	fmt.Println("Doc Serials:", cDocSeries)
	fmt.Println("Cartoons:", cCartoonMovie)
	fmt.Println("Cartoons Serial:", cCartoonSeries)
	fmt.Println("TV  Show:", cTVShow)
	fmt.Println("Animes:", cAnime)

	fmt.Println("Torrents with IMDB:", wIMDB)
	fmt.Println("Torrents without IMDB:", nIMDB)
	fmt.Println("Torrents:", len(listTorr))
}

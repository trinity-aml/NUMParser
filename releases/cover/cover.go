package cover

import (
	"errors"
	"image"
	"image/color"
	"log"
	"net/http"

	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

const cardW, cardH = 267, 400  // размер постеров
const width, height = 748, 440 // размер хоста

func drawCard(dc *gg.Context, img image.Image, x, y, angle, darken float64) {
	// отдельный контекст для постера
	cardDC := gg.NewContext(cardW, cardH)

	// закругляем углы и клип
	cardDC.DrawRoundedRectangle(0, 0, float64(cardW), float64(cardH), 20)
	cardDC.Clip()

	// масштабирование под размер карточки, сохраняя пропорции
	iw := float64(img.Bounds().Dx())
	ih := float64(img.Bounds().Dy())
	scale := float64(cardW) / iw
	if ih*scale > float64(cardH) {
		scale = float64(cardH) / ih
	}

	cardDC.ScaleAbout(scale, scale, float64(cardW)/2, float64(cardH)/2)

	// рисуем изображение в центр
	cardDC.DrawImageAnchored(img, cardW/2, cardH/2, 0.5, 0.5)

	// затемнение (делаем после DrawImageAnchored, чтобы оно было на всю карточку)
	if darken > 0 {
		cardDC.SetRGBA(0, 0, 0, darken)
		cardDC.DrawRectangle(-cardW/2, -cardH/2, float64(cardW*2), float64(cardH*2))
		cardDC.Fill()
	}

	// переносим готовую карточку на главный холст
	dc.Push()
	dc.Translate(x+float64(cardW)/2, y+float64(cardH)/2)
	dc.Rotate(gg.Radians(angle))
	dc.DrawImageAnchored(cardDC.Image(), 0, 0, 0.5, 0.5)
	dc.Pop()
}

func BuildCover(urls []string, backs string, filename string) error {
	log.Println("Building cover")
	log.Println("Building posters:")
	for _, url := range urls {
		log.Println(" ", url)
	}
	log.Println("Backdrop:", backs)
	log.Println("Filename:", filename)
	dc := gg.NewContext(width, height)

	if backs != "" {
		resp, err := http.Get(backs)
		if err != nil {
			log.Println("Error fetching backdrop:", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			log.Println("Error fetching backdrop:", resp.StatusCode)
			return errors.New("Error fetching backdrop: " + resp.Status)
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Println("Error decoding back image:", err)
			return err
		}
		// масштабируем фон под холст
		bg := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)
		// затемнение
		bg = imaging.AdjustBrightness(bg, -50) // -0.3 — уменьшение яркости на 30%
		// размытие
		bg = imaging.Blur(bg, 5.0) // радиус размытия 2, можно регулировать
		dc.DrawImage(bg, 0, 0)
	} else {
		// фон — градиент
		grad := gg.NewLinearGradient(0, 0, float64(width), float64(height))
		grad.AddColorStop(0, color.RGBA{38, 38, 38, 255})
		grad.AddColorStop(1, color.RGBA{64, 64, 64, 255})
		dc.SetFillStyle(grad)
		dc.DrawRectangle(0, 0, float64(width), float64(height))
		dc.Fill()
	}

	// загружаем изображения
	var imgs []image.Image
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Println("Error fetching poster:", err)
			return err
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Println("Error decoding poster:", err)
			return err
		}
		imgs = append([]image.Image{img}, imgs...)
	}

	// позиции и углы (снизу вверх)
	offsets := [][]float64{
		{cardW*3 - 150, height - cardH + 120, 15}, // 4
		{cardW*2 - 100, height - cardH + 60, 10},  // 3
		{cardW - 50, height - cardH + 20, 5},      // 2
		{0, height - cardH, 0},                    // 1
	}

	count := len(imgs)
	for i := 0; i < len(imgs); i++ {
		x, y, angle := offsets[i][0], offsets[i][1], offsets[i][2]
		darken := float64(count-i-1) / float64(count) * 0.9
		drawCard(dc, imgs[i], x-50, y+20, angle, darken)
	}

	img := dc.Image()
	imgSmall := imaging.Blur(img, 0.5)
	return imaging.Save(imgSmall, filename)
}

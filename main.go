package main

import (
	"bytes"
	"context"
	"flag"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"

	"image/draw"
	"image/jpeg"
	_ "image/jpeg"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var imageURL string
var s3Region string
var s3Bucket string
var tempDir string
var rect string
var r image.Rectangle

func init() {
	flag.StringVar(&imageURL, "imageURL", "", "The URL of the image to fetch.")
	flag.StringVar(&s3Region, "s3Region", "us-east-1", "The AWS region to use.")
	flag.StringVar(&s3Bucket, "s3Bucket", "", "The S3 bucket name (not the ARN) to upload the snapshots to.")
	flag.StringVar(&tempDir, "tempDir", ".", "The directory in which to store the temporary downloaded images.")
	flag.StringVar(&rect, "rect", "", "The x/y coordinates that define the rectangle to use to crop in the form of x,y,x,y")
}

func main() {

	flag.Parse()

	if rect != "" {
		rpoints := strings.Split(rect, ",")
		if len(rpoints) != 4 {
			log.Fatal("rect flag not in the form of x,y,x,y")
		}
		var ripoints [4]int
		for i, num := range rpoints {
			v, err := strconv.Atoi(num)
			if err != nil {
				log.Fatal("Error converting number " + num + " to int.")
			}
			ripoints[i] = v
		}
		r = image.Rect(ripoints[0], ripoints[1], ripoints[2], ripoints[3])
	}

	s, err := session.NewSession(&aws.Config{Region: aws.String(s3Region)})
	if err != nil {
		log.Fatal("Error initializing AWS Session", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c:
			cancel()
		}
	}()

	log.Println("Running...")
	processSnapshot(s)

	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-ticker.C:
			processSnapshot(s)
		case <-ctx.Done():
			return
		}
	}
}

func processSnapshot(s *session.Session) {

	fn, err := fetchImage()
	if err != nil {
		log.Println("Downoad failed.", err.Error())
	} else {
		if rect != "" {
			err = cropImage(fn, r)
		}
		if err != nil {
			log.Println("Crop failed.", err.Error())
		} else {
			err := uploadImage(s, fn)
			if err != nil {
				log.Println("Error uploading image to s3", err)
			} else {
				log.Println("Upload Successful")
			}
		}
	}
}

func fetchImage() (string, error) {
	url := imageURL

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	fn := tempDir + "/" + time.Now().Format(time.RFC3339) + ".jpeg"
	f, err := os.Create(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		return "", err
	}

	return fn, nil
}

func cropImage(fn string, r image.Rectangle) error {

	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	srcimg, _, err := image.Decode(f)

	memimg := image.NewRGBA(srcimg.Bounds())

	draw.Draw(memimg, memimg.Bounds(), srcimg, image.Point{0, 0}, draw.Src)
	newimg := memimg.SubImage(r)

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	w, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return err
	}

	err = jpeg.Encode(w, newimg, nil)
	return err
}

func uploadImage(s *session.Session, fn string) error {

	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, fileName := path.Split(fn)

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s3Bucket),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err == nil {
		os.Remove(fn)
	}

	return err
}

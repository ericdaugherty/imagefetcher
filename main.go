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
var rect string
var r image.Rectangle

var buf bytes.Buffer

func init() {
	flag.StringVar(&imageURL, "imageURL", "", "The URL of the image to fetch.")
	flag.StringVar(&s3Region, "s3Region", "us-east-1", "The AWS region to use.")
	flag.StringVar(&s3Bucket, "s3Bucket", "", "The S3 bucket name (not the ARN) to upload the snapshots to.")
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

	err := fetchImage()
	if err != nil {
		log.Println("Downoad failed.", err.Error())
	} else {
		if rect != "" {
			err = cropImage(r)
		}
		if err != nil {
			log.Println("Crop failed.", err.Error())
		} else {
			err := uploadImage(s)
			if err != nil {
				log.Println("Error uploading image to s3", err)
			} else {
				log.Println("Upload Successful")
			}
		}
	}
}

func fetchImage() error {
	url := imageURL
	buf.Reset()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(&buf, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func cropImage(r image.Rectangle) error {

	srcimg, _, err := image.Decode(&buf)

	memimg := image.NewRGBA(srcimg.Bounds())

	draw.Draw(memimg, memimg.Bounds(), srcimg, image.Point{0, 0}, draw.Src)
	newimg := memimg.SubImage(r)

	buf.Reset()
	err = jpeg.Encode(&buf, newimg, nil)

	return err
}

func uploadImage(s *session.Session) error {

	fileName := time.Now().Format(time.RFC3339) + ".jpeg"

	len := buf.Len()

	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s3Bucket),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buf.Bytes()),
		ContentLength:        aws.Int64(int64(len)),
		ContentType:          aws.String(http.DetectContentType(buf.Bytes())),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	return err
}

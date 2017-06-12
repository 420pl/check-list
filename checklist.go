package checklist

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	httpProto         = "http://"
	httpsProto        = "https://"
	propPairSeparator = "="
	themeName         = "white"
	embedWidth        = "500"
)

// CheckFile reads HTML file containing playlist from loadPath,
// checks it for inconsistencies and fixes them, if there are any.
// Then, function saves fixed file to savePath.
func CheckFile(loadPath, savePath string) (report []string, err error) {
	doc, err := parseFile(loadPath)
	if err != nil {
		return nil, err
	}

	embed, err := getEmbed(doc)
	if err != nil {
		return nil, err
	}

	report = make([]string, 0, 5)
	fixSrcAttribute(embed, &report)
	fixWidthAttribute(embed, &report)

	writeDocumentToFile(savePath, doc)
	return report, nil
}

func parseFile(path string) (*goquery.Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to open file '%s' (%s)", path, err))
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Invalid file '%s', can't parse (%s)", path, err))
	}
	return doc, nil
}

func writeDocumentToFile(path string, doc *goquery.Document) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to create file '%s' (%s)", path, err))
	}
	defer file.Close()

	file.WriteString(doc.Text())
	html, err := goquery.OuterHtml(doc.Find("embed"))
	if err != nil {
		return errors.New(fmt.Sprintf("Unexpected error: %s", err))
	}

	// We need this hack because for whatever reason goquery tends to escape special characters,
	// when setting an attribute of element.
	html = strings.Replace(html, "&amp;", "&", -1)
	file.WriteString(html)

	return nil
}

func getEmbed(doc *goquery.Document) (*goquery.Selection, error) {
	embedNode := doc.Find("embed")
	switch embedNode.Size() {
	case 0:
		return nil, errors.New("Required <embed> element not found")
	case 1:
		return embedNode, nil
	default:
		return nil, errors.New("Found multiple <embed> elements")
	}
}

func fixSrcAttribute(embedNode *goquery.Selection, finalReport *[]string) error {
	srcAttr, isSrcExist := embedNode.Attr("src")
	if !isSrcExist {
		return errors.New("Required 'src' attribute of <embed> element not found")
	}

	srcSplitted := strings.Split(srcAttr, "?")
	fixedSrc := fixPlaylistUrl(srcSplitted[0], finalReport) + "?" + fixPlaylistProperties(srcSplitted[1], finalReport)
	embedNode.SetAttr("src", fixedSrc)
	return nil
}

func fixPlaylistUrl(url string, report *[]string) string {
	if strings.HasPrefix(url, httpProto) {
		*report = append(*report, "Malformed playlist URL - http:// used instead of https://")
		return strings.Replace(url, httpProto, httpsProto, 1)
	}
	return url
}

func fixPlaylistProperties(rawProps string, report *[]string) string {
	srcPropPairs := strings.Split(rawProps, "&")
	fixedProps := make([]string, len(srcPropPairs))
	for i, prop := range srcPropPairs {
		propPair := strings.Split(prop, propPairSeparator)
		switch propPair[0] {
		case "theme":
			if propPair[1] != themeName {
				*report = append(*report, "Invalid playlist theme ('theme' property) - should be white")
				propPair[1] = themeName
				prop = strings.Join(propPair, propPairSeparator)
			}
		case "w":
			if propPair[1] != embedWidth {
				*report = append(*report, "Invalid width value ('w' property) - should be 500")
				propPair[1] = embedWidth
				prop = strings.Join(propPair, propPairSeparator)
			}
		case "withart":
			if strings.HasPrefix(propPair[1], httpProto) {
				*report = append(*report, "Malformed cover art URL ('withart' property)- http:// used instead of https://")
				prop = strings.Replace(prop, httpProto, httpsProto, 1)
			}
		}
		fixedProps[i] = prop
	}
	return strings.Join(fixedProps, "&")
}

func fixWidthAttribute(embedNode *goquery.Selection, report *[]string) error {
	widthAttr, isWidthExist := embedNode.Attr("width")
	if !isWidthExist {
		return errors.New("Required 'width' attribute of <embed> element not found")
	}

	if widthAttr != embedWidth {
		*report = append(*report, "Invalid width value - should be 500")
		embedNode.SetAttr("width", embedWidth)
	}
	return nil
}

package ntomd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/dstotijn/go-notion"
)

func GetMarkdownStringFromNotionPage(notionKey, pageId *string) (*string, error) {
	if notionKey == nil || pageId == nil {
		return nil, errors.New("'notionKey' and 'pageId' are both required")
	}

	client := notion.NewClient(*notionKey)

	page, err := client.FindPageByID(context.Background(), *pageId)
	if err != nil {
		return nil, fmt.Errorf("(GetMarkdownStringFromNotionPage) client.FindPageById: %v", err)
	}

	hasMeta := false
	meta := "---\n"

	body := ""

	props := page.Properties.(notion.DatabasePageProperties)
	for k, v := range props {
		// log.Println(k, v.Type)
		pstr, err := ParseProperty(v)
		if err != nil {
			log.Fatal(err)
		}
		if pstr != nil {
			hasMeta = true
			meta += fmt.Sprintf("%v: %v\n", k, *pstr)
		}
	}
	meta += "---\n"

	children, err := client.FindBlockChildrenByID(context.Background(), *pageId, nil)
	if err != nil {
		return nil, fmt.Errorf("(GetMarkdownStringFromNotionPage) client.FindBlockChildrenByID: %v", err)
	}

	for _, el := range children.Results {
		switch block := el.(type) {
		case *notion.ParagraphBlock:
			bstr, err := ParseRichText(block.RichText)
			if err != nil {
				return nil, err
			}
			if *bstr != "" {
				body += *bstr + "\n\n"
			}
		case *notion.Heading1Block:
			bstr, err := ParseRichText(block.RichText)
			if err != nil {
				return nil, err
			}
			body += fmt.Sprintf("# %v\n\n", *bstr)
		case *notion.Heading2Block:
			bstr, err := ParseRichText(block.RichText)
			if err != nil {
				return nil, err
			}
			body += fmt.Sprintf("## %v\n\n", *bstr)
		case *notion.Heading3Block:
			bstr, err := ParseRichText(block.RichText)
			if err != nil {
				return nil, err
			}
			body += fmt.Sprintf("### %v\n\n", *bstr)
		case *notion.DividerBlock:
			body += "---\n\n"
		case *notion.ImageBlock:
			bstr, err := ParseImageBlock(block)
			if err != nil {
				return nil, err
			}
			body += *bstr + "\n\n"
		default:
			log.Println("TRIED TO PARSE UNHANDLED BLOCK!!!", reflect.TypeOf(block))
		}
	}

	if hasMeta {
		body = fmt.Sprintf("%v\n%v", meta, body)
	}

	return &body, nil
}

func ParseProperty(prop notion.DatabasePageProperty) (*string, error) {
	if prop.Type == "rich_text" || prop.Type == "name" {
		return ParseRichText(prop.RichText)
	}

	if prop.Type == "url" {
		return prop.URL, nil
	}
	return nil, nil
}

func ParseRichText(richText []notion.RichText) (*string, error) {
	str := ""
	for _, el := range richText {
		stri := ""
		if el.Annotations.Bold {
			str += "**"
		}
		if el.Annotations.Italic {
			stri += "_"
		}
		stri += el.Text.Content
		if el.Annotations.Italic {
			stri += "_"
		}
		if el.Annotations.Bold {
			stri += "**"
		}
		str += stri
	}
	return &str, nil
}

func ParseImageBlock(block *notion.ImageBlock) (*string, error) {
	path := ""
	if block.File != nil {
		path = block.File.URL
	}
	if block.External != nil {
		path = block.External.URL
	}
	if path == "" {
		return nil, nil
	}
	caption, err := ParseRichText(block.Caption)
	if err != nil {
		return nil, err
	}
	str := fmt.Sprintf("![%v](%v)", *caption, path)
	return &str, nil
}

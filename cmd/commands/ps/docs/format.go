package docs

import (
	"bmoc/cmd/utils"
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forPelevin/gomoji"
	slugify "github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

var (
	formatDocFlag string
	noCleanupFlag bool
)

// PsCmd represents the ps command
var FormatCmd = &cobra.Command{
	Use:   "format",
	Short: "short description",
	Run: func(cmd *cobra.Command, args []string) {
		if formatDocFlag == "" {
			log.Println("Missing parameter, 'in (i)' is required")
			return
		}
		FormatDoc(formatDocFlag, noCleanupFlag)
	},
}

func init() {
	FormatCmd.Flags().StringVarP(&formatDocFlag, "in", "i", "", "Extracts a Notion exported zip & processes for the docs site.")
	FormatCmd.Flags().BoolVar(&noCleanupFlag, "no-cleanup", false, "If set, the original & generated temp files will not be deleted.")
}

func FormatDoc(zipPath string, noCleanupFlag bool) {
	outpath := ""
	imgoutpath := ""
	path, err := utils.UnzipSource(zipPath)
	if err != nil {
		log.Fatal(err)
	}

	// Read in the doc
	fileArr, err := utils.WalkMatch(*path, "*.md")
	if err != nil {
		log.Fatal(err)
	}

	if len(fileArr) == 1 {
		title := ""
		subtitle := ""
		slug := ""
		content := ""

		dat, err := os.ReadFile(fileArr[0])
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(strings.NewReader(string(dat)))
		skipLine := ""
		isBuildingInfoBlock := false
		isInChecklist := false
		// infoBlockType := "note"
		infoBlockTypeWasCaptured := false
		for scanner.Scan() {
			line := scanner.Text()

			// Get the title of the post, create necessary directories
			if strings.HasPrefix(line, "# ") {
				line = strings.Replace(line, "# ", "", 1)
				title = line
				slug = slugify.Make(line)
				myself, err := user.Current()
				if err != nil {
					log.Fatal(err)
				}

				outpath = fmt.Sprintf("%v/Desktop/%v", myself.HomeDir, slug)
				outpath, err = filepath.Abs(outpath)
				if err != nil {
					log.Fatal(err)
				}

				utils.MkDir(outpath)
				imgoutpath = fmt.Sprintf("%v/%v", outpath, slug)
				imgoutpath, err = filepath.Abs(imgoutpath)
				if err != nil {
					log.Fatal(err)
				}

				utils.MkDir(imgoutpath)
				continue
			}

			// Extract the excerpt
			if strings.HasPrefix(line, "+Excerpt: ") {
				line = strings.Replace(line, "+Excerpt: ", "", 1)
				subtitle = line
				continue
			}

			if strings.HasPrefix(line, "**Checklist") {
				isInChecklist = true
				continue
			}

			if isInChecklist && line != "---" {
				continue
			}

			if isInChecklist && line == "---" {
				isInChecklist = false
				continue
			}

			// TODO: this will have to be implemented when I hit the API directly
			// Inline block links have the hash values removed, so I cant pull this off
			// Replace Notion block links
			// r, err := regexp.Compile(`\((.*)\)`)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// found := r.FindAllString(line, 1)
			// for _, el := range found {
			// 	str := strings.Replace(el, "(", "", 1)
			// 	str = strings.Replace(str, ")", "", 1)
			// 	if strings.HasPrefix(str, "https://www.notion.so") {
			// 		href := services.GetHeaderBlockText(str)
			// 		if href != nil {
			// 			content = strings.Replace(content, el, *href, 1)
			// 		}
			// 	}
			// }

			// Handle images
			if strings.HasPrefix(line, "![") {
				imgPath := ""
				imgSlug := ""
				imgAlt := ""
				imgExt := ""

				// get title/alt of the image
				r, err := regexp.Compile(`\!\[(.*)\]`)
				if err != nil {
					log.Fatal(err)
				}

				found := r.FindAllString(line, 1)
				if len(found) == 1 {
					imgAlt = found[0]
					imgAlt = strings.Replace(imgAlt, "![", "", 1)
					imgAlt = strings.Replace(imgAlt, "]", "", 1)
					if strings.HasPrefix(imgAlt, "capture_") || strings.HasPrefix(imgAlt, "Untitled") {
						log.Println(fmt.Sprintf("WARN: Found possible uncaptioned image: %v", imgAlt))
					}
					imgSlug = slugify.Make(imgAlt)
				}

				// Get the path of the image
				r, err = regexp.Compile(`\((.*)\)`)
				if err != nil {
					log.Fatal(err)
				}

				found = r.FindAllString(line, 1)
				if len(found) == 1 {
					imgPath = found[0]
					imgPath = strings.Replace(imgPath, "(", "", 1)
					imgPath = strings.Replace(imgPath, ")", "", 1)
					imgPath = strings.ReplaceAll(imgPath, "%20", " ")
					imgPath = *path + "/" + imgPath

					splitPath := strings.Split(found[0], ".")
					imgExt = splitPath[len(splitPath)-1]
					imgExt = strings.Replace(imgExt, ")", "", 1)

					_, err := utils.Copy(imgPath, fmt.Sprintf("%v/%v.%v", imgoutpath, imgSlug, imgExt))
					if err != nil {
						log.Fatal(err)
					}
				}

				content += fmt.Sprintf("![%v](/docs/img/docs/%v/%v.%v)", imgAlt, slug, imgSlug, imgExt)
				skipLine = imgAlt
				continue
			}

			// Images have the same title on the following line, so dont put it into content
			if line == skipLine {
				skipLine = ""
				continue
			}

			if line == "<aside>" {
				isBuildingInfoBlock = true
				content += "<InfoBlock"
				continue
			}

			if isBuildingInfoBlock && !infoBlockTypeWasCaptured {
				// emoji := string(line[0])
				emoji := gomoji.FindAll(line)
				if len(emoji) > 0 && emoji[0].Character == "⚠️" {
					content += " type=\"warning\">\n\n"
					content += fmt.Sprintf("%v\n\n", line[6:len(line)-1])
				} else {
					content += " type=\"note\">\n\n"
					content += fmt.Sprintf("%v\n\n", line[5:len(line)-1])
				}
				infoBlockTypeWasCaptured = true
				continue
			}

			if line == "</aside>" {
				isBuildingInfoBlock = false
				infoBlockTypeWasCaptured = false
				// infoBlockType = "note"
				content += "</InfoBlock>\n\n"
				continue
			}

			// Standard content
			content += fmt.Sprintf("%v\n\n", line)
		}

		content += "<GetHelp />\n"

		// Write the content out
		outcontent := fmt.Sprintf("---\ntitle: %v\nsubtitle: %v\n---\n\n%v", title, subtitle, content)
		outfile := fmt.Sprintf("%v/index.mdx", outpath)
		err = os.WriteFile(outfile, []byte(outcontent), 0644)
		if err != nil {
			log.Fatal(err)
		}

		// Write the JSON meta out
		outjson := fmt.Sprintf("{\n\t\"display\": \"%v\",\n\t\"route\": \"%v\"\n}", title, slug)
		outjsonfile := fmt.Sprintf("%v/meta.json", outpath)
		err = os.WriteFile(outjsonfile, []byte(outjson), 0644)
		if err != nil {
			log.Fatal(err)
		}

		//Cleanup
		if !noCleanupFlag {
			err = os.Remove(zipPath)
			if err != nil {
				log.Fatal(err)
			}

			err = os.RemoveAll(*path)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

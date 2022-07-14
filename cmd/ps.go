/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	slugify "github.com/gosimple/slug"
)

var (
	formatDocFlag string
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "PlanetScale stuffz",
	Run:   run,
}

func init() {
	rootCmd.AddCommand(psCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// psCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	psCmd.Flags().StringVarP(&formatDocFlag, "format-doc", "d", "", "Help message for toggle")
}

func run(cmd *cobra.Command, args []string) {
	if formatDocFlag != "" {
		// Unzip the thing
		outpath := ""
		imgoutpath := ""
		path, err := unzipSource(formatDocFlag)
		if err != nil {
			log.Fatal(err)
		}

		// Read in the doc
		fileArr, err := walkMatch(*path, "*.md")
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
			for scanner.Scan() {
				line := scanner.Text()

				// Get the title of the post, create necessary directories
				if strings.HasPrefix(line, "#") {
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

					mkDir(outpath)
					imgoutpath = fmt.Sprintf("%v/%v", outpath, slug)
					imgoutpath, err = filepath.Abs(imgoutpath)
					if err != nil {
						log.Fatal(err)
					}

					mkDir(imgoutpath)
					continue
				}

				// Extract the excerpt
				if strings.HasPrefix(line, "+Excerpt: ") {
					line = strings.Replace(line, "+Excerpt: ", "", 1)
					subtitle = line
					continue
				}

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

						_, err := copy(imgPath, fmt.Sprintf("%v/%v.%v", imgoutpath, imgSlug, imgExt))
						if err != nil {
							log.Fatal(err)
						}
					}

					content += fmt.Sprintf("![%v](/img/docs/%v/%v.%v)", imgAlt, slug, imgSlug, imgExt)
					skipLine = imgAlt
					continue
				}

				// Images have the same title on the following line, so dont put it into content
				if line == skipLine {
					skipLine = ""
					continue
				}

				// put the content in
				content += fmt.Sprintf("%v\n\n", line)
			}

			outcontent := fmt.Sprintf("---\ntitle: %v\nsubtitle: %v\n---\n\n%v", title, subtitle, content)
			outfile := fmt.Sprintf("%v/index.mdx", outpath)
			err = os.WriteFile(outfile, []byte(outcontent), 0644)
			if err != nil {
				log.Fatal(err)
			}

			outjson := fmt.Sprintf("{\n\t\"display\": \"%v\",\n\t\"route\": \"%v\"\n}", title, slug)
			outjsonfile := fmt.Sprintf("%v/meta.json", outpath)
			err = os.WriteFile(outjsonfile, []byte(outjson), 0644)
			if err != nil {
				log.Fatal(err)
			}

			//Cleanup
			err = os.Remove(formatDocFlag)
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

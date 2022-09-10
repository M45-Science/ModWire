package support

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"../glob"
	embed "github.com/clinet/discordgo-embed"
	"github.com/mitchellh/go-wordwrap"
)

func SendModInfo(res glob.ModDB, update_message int, channel string) {
	//-------------------------
	//-- CHANGELOG CLEANING --
	//-------------------------
	//Get changelog
	changelog_string := res.Changelog

	//No changelog? grr... we'll use description instead.
	if len(changelog_string) < 8 {
		changelog_string = res.Description
	}

	//In case of empty changelog/description
	if len(changelog_string) < 8 {
		changelog_string = res.Summary
	}

	//Sneaky way to fix line endings?
	changelog_string = strings.ReplaceAll(changelog_string, "\r", "\n")

	changelog_string = strings.ReplaceAll(changelog_string, "\n\n\n", "\n")
	changelog_string = strings.ReplaceAll(changelog_string, "\n\n", "\n")

	changelog_string = strings.ReplaceAll(changelog_string, " \n", "\n")
	changelog_string = strings.ReplaceAll(changelog_string, "\n ", "\n")

	//Remove extra stuff
	changelog_string = strings.ReplaceAll(changelog_string, "-", "")  //Dash lines
	changelog_string = strings.ReplaceAll(changelog_string, "_", "")  //Underscore lines
	changelog_string = strings.ReplaceAll(changelog_string, "\t", "") //Tabs

	//Split text, into array of lines.
	changelog_lines := strings.SplitAfter(changelog_string, "\n")
	changelog_numlines := len(changelog_lines)

	//Word wrap
	for curline := 0; curline < changelog_numlines; curline++ {
		//remove, then add back the newline for filtering
		changelog_lines[curline] = strings.ReplaceAll(changelog_lines[curline], "\n", "")

		//Split into words for filtering
		words := strings.Split(changelog_lines[curline], " ")
		numwords := len(words)

		//Remove annoyances
		for w := 0; w < numwords; w++ {
			//Remove URLs
			if strings.Contains(strings.ToLower(words[w]), "patreon") ||
				strings.Contains(strings.ToLower(words[w]), "discord") ||
				strings.Contains(strings.ToLower(words[w]), "://") {
				if glob.XDEBUG {
					//Log("Removing url: ", words[w])
				}
				words[w] = "*removed*" //Remove word
			}
		}

		//Reassemble
		changelog_lines[curline] = strings.Join(words, " ") + "\n"

		//Word wrap last
		mll, _ := strconv.ParseUint(Config.DiscordMaxLineLength, 10, 16)
		changelog_lines[curline] = wordwrap.WrapString(changelog_lines[curline], uint(mll))

	}

	cutlines := false
	//Convert to normal string
	changelog_string_b := strings.Join(changelog_lines[:changelog_numlines], "")

	//Split text, into array of lines... again.
	changelog_lines_b := strings.SplitAfter(changelog_string_b, "\n")
	changelog_numlines = len(changelog_lines_b)

	//Cut new number of lines down
	mll, _ := strconv.Atoi(Config.DiscordMaxLines)
	if changelog_numlines > mll {
		changelog_numlines = mll
		cutlines = true
	}

	//Convert to single string again
	changelog_line_c := strings.Join(changelog_lines_b[:changelog_numlines], "")

	//----------------------------
	//-- END CHANGELOG CLEANING --
	//----------------------------

	//Setup thumnail/titles/links
	thumb := fmt.Sprintf("http://bhmm.net/images/fact-noimage2.png")

	//Shorten Title
	//Remove newlines/tabs/doublespace
	temp_title := strings.ReplaceAll(res.Title, "\n", "")
	temp_title = strings.ReplaceAll(temp_title, "\r", "")
	temp_title = strings.ReplaceAll(temp_title, "\t", "")
	temp_title = strings.ReplaceAll(temp_title, "  ", " ")

	title_words := strings.Split(temp_title, " ")
	title_words_count := len(title_words)

	mtw, _ := strconv.Atoi(Config.DiscordMaxTitleWords)
	//Shorten if long, add elipse to indicate cut
	if title_words_count > mtw {
		title_words_count = mtw
		title_words[title_words_count-1] = "..."
	}

	//Recombine
	cleaned_title := strings.Join(title_words[:title_words_count], " ")

	title := fmt.Sprintf("**%s** (%s)", cleaned_title,
		res.Latest.Version)

	//Replace wube's 404 'no thumbnail' image.
	if res.Thumbnail != "/assets/.thumb.png" {
		thumb = fmt.Sprintf("https://mods-data.factorio.com%s", res.Thumbnail)
	}

	//Factorio icon in title
	//This should eventually be set by game catagory
	image := fmt.Sprintf("http://bhmm.net/images/gear-small.png")

	//Changelog and link to read more
	buffer := ""

	cleanurl := &url.URL{Path: res.Name}
	cleanurlstring := cleanurl.String()

	if cutlines {
		buffer = fmt.Sprintf("(by [%s](https://mods.factorio.com/user/%s), %d [downloads](https://mods.factorio.com%s))\n%s[Read more...](https://mods.factorio.com/mod/%s/changelog)",
			res.Owner, res.Owner, res.Downloads_Count,
			res.Latest.Download_Url, changelog_line_c,
			cleanurlstring)
	} else {
		buffer = fmt.Sprintf("(by [%s](https://mods.factorio.com/user/%s), %d [downloads](https://mods.factorio.com%s))\n%s",
			res.Owner, res.Owner, res.Downloads_Count,
			res.Latest.Download_Url, changelog_line_c)
	}

	//Change title for update
	upstring := "Factorio Mod:\n\r\n\r"
	if update_message == glob.TypeUpdate {
		upstring = "Factorio Mod Update:\n\r\n\r"
	} else if update_message == glob.TypeDelete {
		upstring = "Factorio Mod Deleted:\n\r\n\r"
	} else if update_message == glob.TypeNew {
		upstring = "NEW Factorio Mod:\n\r\n\r"
	}

	//Slap data into embed format.
	myembed := embed.NewEmbed().
		SetTitle(title).
		SetThumbnail(thumb).
		SetDescription(buffer).
		SetAuthor(upstring, image).
		SetURL(fmt.Sprintf("https://mods.factorio.com/mod/%s", cleanurlstring)).
		SetColor(0xff0000).MessageEmbed

	//Send it off!
	_, err := glob.DS.ChannelMessageSendEmbed(channel, myembed)
	if err != nil {
		Log("Couldn't send message: stat command")
	}

}

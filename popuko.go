// made by amy (amyadzuki@gmail.com)
// this file is released into the public domain

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	///////////////////////////////////////////////////////////////////////////////
	////////////////////////  BEGIN CUSTOMIZATION SECTION  ////////////////////////
	///////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////
	// ABOUT YOU AND YOUR SERVER; PLEASE DON'T PUT NUMBERS IN QUOTATION MARKS  ////
	ClientID  = 442381838611120148
	OwnerID   = 77256980288253952
	BotUserID = 442381838611120148
	ApplyGID  = 378599231583289346
	ApplyCID  = 440399561542991872 // TODO
	ResultGID = 339435151501033474
	ResultCID = 447641375274434560

	///////////////////////////////////////////////////////////////////////////////
	// ABOUT YOU AND YOUR SERVER; THESE ARE TEXT AND DO GO IN QUOTATION MARKS  ////
	Token    = "YOUR TOKEN GOES HERE"
	Playing  = "\u2022 DM me to apply for partnership."
	Password = "LLent Poppy Pappy Day"

///////////////////////////////////////////////////////////////////////////////
/////////////////////////  END CUSTOMIZATION SECTION  /////////////////////////
///////////////////////////////////////////////////////////////////////////////
)

type State struct {
	Count  string
	Role   string
	Desc   string
	Why    string
	Note   string
	State  uint64
	Invite *discordgo.Invite
}

var States map[int64]State
var PRG *rand.Rand
var Mutex sync.Mutex
var StatesMutex sync.Mutex
var CliID string
var OwnID string
var BotID string
var ApGID string
var ApCID string
var RsGID string
var RsCID string
var Passwd string
var REM, RELT *regexp.Regexp

func init() {
	States = make(map[int64]State)
	PRG = rand.New(rand.NewSource(time.Now().UnixNano()))
	CliID = strconv.FormatInt(ClientID, 10)
	OwnID = strconv.FormatInt(OwnerID, 10)
	BotID = strconv.FormatInt(BotUserID, 10)
	ApGID = strconv.FormatInt(ApplyGID, 10)
	ApCID = strconv.FormatInt(ApplyCID, 10)
	RsGID = strconv.FormatInt(ResultGID, 10)
	RsCID = strconv.FormatInt(ResultCID, 10)
	Passwd = strings.ToLower(Password)
	REM = regexp.MustCompile(`^<@!?` + BotID + `>\s*(.*)$`)
	RELT = regexp.MustCompile(`(^|[^\\])<`)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	fmt.Println("This bot is ©2018 amy (amyadzuki@gmail.com).")
	tp := flag.String("t", Token, "Bot token")
	flag.Parse()
	t := *tp
	dg, err := discordgo.New("Bot " + t)
	check(err)
	dg.AddHandler(onMessageCreate)
	check(dg.Open())
	defer dg.Close()
	fmt.Println("======= BOT UP (type Ctrl-C to exit)")
	defer fmt.Println("\n....... BOT DOWN")
	dg.UpdateStatus(0, Playing)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func onMessageCreate(session *discordgo.Session, arg *discordgo.MessageCreate) {
	message := arg.Message

	// Don't scan the bot's own messages:
	if message.Author.ID == BotID {
		return
	}

	// Query information about the channel
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		// Do it the slow but reliable way if the quick but unreliable way fails
		channel, err = session.Channel(message.ChannelID)
		if err != nil {
			return // Give up if the reliable way fails
		}
	}

	if channel.Type != discordgo.ChannelTypeDM {
		return // Ignore any message that is not a DM
	}

	uid := message.Author.ID
	u64, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return
	}

	Mutex.Lock()
	defer Mutex.Unlock()

	state, ok := States[u64]
	if !ok {
		AtPass(session, message, u64)
		return
	}
	switch state.State & 0xf {
	case 0x0:
		Invite(session, message, u64)
	case 0x1:
		MemberCount(session, message, u64)
	case 0x2:
		YourRole(session, message, u64)
	case 0x3:
		Description(session, message, u64)
	case 0x4:
		WhyPartner(session, message, u64)
	case 0x5:
		Comments(session, message, u64)
	case 0x6:
		Submit(session, message, u64)
	case 0xf:
		AtPass(session, message, u64)
	default:
		SomethingWentWrong(session, message, u64)
	}
}

func GoTime(session *discordgo.Session, message *discordgo.Message, u64 int64, set uint8) {
	set_lower := uint64(set) & 0xf
	set_upper := uint64(set) & ^uint64(0xf)

	time.Sleep(10 * time.Minute)

	StatesMutex.Lock()
	defer StatesMutex.Unlock()

	state, ok := States[u64]
	if !ok {
		return
	}

	sta_lower := state.State & 0xf
	sta_upper := state.State & ^uint64(0xf)

	if sta_upper == set_upper { // still same instance
		if sta_lower <= set_lower { // no progress made
			delete(States, u64)
			if set_lower != 0xf {
				session.ChannelMessageSend(message.ChannelID,
					"Your time has ran out. Enter the password to "+
						"initiate the application process again.\n"+
						"https://i.imgur.com/MKpQA9a.gif")
			} // otherwise it was a success
		}
	}
}

func InitAndTime(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	States[u64] = State{State: PRG.Uint64() << 4}
	go GoTime(session, message, u64, 0)
}

func SetAndTime(session *discordgo.Session, message *discordgo.Message, u64 int64, set uint8) {
	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	state, _ := States[u64] // TODO: really don't need `ok`?
	state.State &= ^uint64(0xf)
	state.State |= uint64(set)
	States[u64] = state
	go GoTime(session, message, u64, set)
}

func SomethingWentWrong(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	session.ChannelMessageSend(message.ChannelID, "Something went wrong.")
	delete(States, u64)
}

func AtPass(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if strings.ToLower(message.Content) == Passwd {
		session.ChannelMessageSend(message.ChannelID,
			"**You have read our <#423168074448109588> Partnership "+
				"section and read our requirements, yes or no? (y/n)**")
		InitAndTime(session, message, u64)
	} else {
		session.ChannelMessageSend(message.ChannelID,
			"Hey! Thanks for DMing me, to start the partnership process, "+
				"do as the <#423168074448109588> and enter the password.\n"+
				"https://i.imgur.com/VyPYpNq.gif")
	}
}

func Invite(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if TooLong(session, message, 4) {
		return
	}
	if message.Content[0] != 'y' && message.Content[0] != 'Y' {
		session.ChannelMessageSend(message.ChannelID,
			"Alright, application process closed. Self-Destruction, activated. "+
				"See ya in Heaven *or Hell*\n"+
				"https://i.imgur.com/lUAXOyD.gif")
		return
	}
	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	state, ok := States[u64]
	if !ok { // race condition
		session.ChannelMessageSend(message.ChannelID,
			"An extremely rare error occurred! Please start again.")
		return
	}
	state.State &= ^uint64(0xf)
	state.State |= uint64(0x1)
	States[u64] = state
	go GoTime(session, message, u64, 0x1)
	session.ChannelMessageSend(message.ChannelID,
		"Thank you for reading our partnership section. "+
			"Nice. Let's initiate the process.\n"+
			"*Please don't take longer than 10 Minutes "+
			"without responding to a question.*\n\n"+
			"**What is your server's permanent invite link?**\n"+
			"https://i.imgur.com/fjKdBc6.gif")
}

var REInvite *regexp.Regexp

func init() {
	REInvite = regexp.MustCompile(`(?i)^(?:(?:https?://)?discord\.gg/)?([^/]+)/?$`)
}
func MemberCount(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if TooLong(session, message, 32) {
		return
	}
	match := REInvite.FindStringSubmatch(message.Content)
	if len(match) != 2 || len(match[1]) < 3 {
		session.ChannelMessageSend(message.ChannelID,
			"There is an error with the invite, make sure it's functioning, "+
				"permanent and unlimited then try again. I will be spinning here "+
				"while you do that."+
				"\n"+"https://i.imgur.com/HM9vsJO.gif")
		return
	}

	invite, err := session.Invite(match[1])
	if err != nil || invite.Guild == nil {
		session.ChannelMessageSend(message.ChannelID,
			"There is an error with the invite, make sure it's functioning, "+
				"permanent and unlimited then try again. I will be spinning here "+
				"while you do that."+
				" Error: Discord says that is not an invite link at all, even "+
				"though it looks like one to me."+
				"\n"+"https://i.imgur.com/HM9vsJO.gif")
		return
	}

	if invite.MaxAge != 0 {
		session.ChannelMessageSend(message.ChannelID,
			"There is an error with the invite, make sure it's functioning, "+
				"permanent and unlimited then try again. I will be spinning here "+
				"while you do that."+
				" Error: It looks like your invite will expire after a while."+
				"\n"+"https://i.imgur.com/HM9vsJO.gif")
		return
	}

	if invite.Uses != 0 {
		session.ChannelMessageSend(message.ChannelID,
			"There is an error with the invite, make sure it's functioning, "+
				"permanent and unlimited then try again. I will be spinning here "+
				"while you do that."+
				" Error: It looks like your invite will expire after a number of uses."+
				"\n"+"https://i.imgur.com/HM9vsJO.gif")
		return
	}

	if invite.Revoked {
		session.ChannelMessageSend(message.ChannelID,
			"There is an error with the invite, make sure it's functioning, "+
				"permanent and unlimited then try again. I will be spinning here "+
				"while you do that."+
				" Error: It looks like somebody manually revoked your invite!"+
				"\n"+"https://i.imgur.com/HM9vsJO.gif")
		return
	}

	//  if invite.Guild.MemberCount < 10 {
	//      session.ChannelMessageSend (message.ChannelID,
	//              "You must have at least 10 members to request a partnership.")
	//      return
	//  }

	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	state, ok := States[u64]
	if !ok { // race condition
		session.ChannelMessageSend(message.ChannelID,
			"An extremely rare error occurred! Please start again.")
		return
	}
	state.State &= ^uint64(0xf)
	state.State |= uint64(0x2)
	state.Invite = invite
	States[u64] = state
	go GoTime(session, message, u64, 0x2)
	session.ChannelMessageSend(message.ChannelID,
		"**How many members does your server have?**")
}

func YourRole(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if TooLong(session, message, 6) {
		return
	}
	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	state, ok := States[u64]
	if !ok { // race condition
		session.ChannelMessageSend(message.ChannelID,
			"An extremely rare error occurred! Please start again.")
		return
	}
	state.State &= ^uint64(0xf)
	state.State |= uint64(0x3)
	state.Count = message.Content
	States[u64] = state
	go GoTime(session, message, u64, 0x3)
	session.ChannelMessageSend(message.ChannelID,
		"**What is your role on your server? "+
			"(Owner, Co-Owner, Partnership Director etc).**")
}

func Description(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if TooLong(session, message, 100) {
		return
	}
	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	state, ok := States[u64]
	if !ok { // race condition
		session.ChannelMessageSend(message.ChannelID,
			"An extremely rare error occurred! Please start again.")
		return
	}
	state.State &= ^uint64(0xf)
	state.State |= uint64(0x4)
	state.Role = message.Content
	States[u64] = state
	go GoTime(session, message, u64, 0x4)
	session.ChannelMessageSend(message.ChannelID,
		"**Tell us your server's description:**\n"+
			"*Keep it under 500 characters.*")
}

func WhyPartner(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if TooLong(session, message, 500) {
		return
	}
	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	state, ok := States[u64]
	if !ok { // race condition
		session.ChannelMessageSend(message.ChannelID,
			"An extremely rare error occurred! Please start again.")
		return
	}
	state.State &= ^uint64(0xf)
	state.State |= uint64(0x5)
	state.Desc = message.Content
	States[u64] = state
	go GoTime(session, message, u64, 0x5)
	session.ChannelMessageSend(message.ChannelID,
		"**Why should we partner with you?**")
}

func Comments(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if TooLong(session, message, 500) {
		return
	}
	StatesMutex.Lock()
	defer StatesMutex.Unlock()
	state, ok := States[u64]
	if !ok { // race condition
		session.ChannelMessageSend(message.ChannelID,
			"An extremely rare error occurred! Please start again.")
		return
	}
	state.State &= ^uint64(0xf)
	state.State |= uint64(0x6)
	state.Why = message.Content
	States[u64] = state
	go GoTime(session, message, u64, 0x6)
	session.ChannelMessageSend(message.ChannelID,
		"**Any additional information or comment that you would like to add?**")
}

func Submit(session *discordgo.Session, message *discordgo.Message, u64 int64) {
	if TooLong(session, message, 500) {
		return
	}
	send := new(discordgo.MessageSend)
	send.Embed = new(discordgo.MessageEmbed)
	send.Embed.Thumbnail = new(discordgo.MessageEmbedThumbnail)
	send.Embed.Fields = make([]*discordgo.MessageEmbedField, 9, 9)
	send.Embed.Fields[0] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[1] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[2] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[3] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[4] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[5] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[6] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[7] = new(discordgo.MessageEmbedField)
	send.Embed.Fields[8] = new(discordgo.MessageEmbedField)

	StatesMutex.Lock() // TODO: make this a function
	state, ok := States[u64]
	if !ok { // race condition
		session.ChannelMessageSend(message.ChannelID,
			"An extremely rare error occurred! Please start again.")
		StatesMutex.Unlock()
		return
	}
	delete(States, u64)
	StatesMutex.Unlock()

	//  send.Content = "https://discord.gg/" + state.Invite.Code
	send.Embed.Title = "\U0001f4e8 New Partnership Application!"
	send.Embed.Description =
		"A new partnership application has been submitted by this user!"
	send.Embed.Color = 3781620 // TODO
	send.Embed.Thumbnail.URL = "https://i.imgur.com/Jg6kDYy.gif"
	send.Embed.Fields[0].Name = "\U0001f539 Username:"
	send.Embed.Fields[0].Value =
		"<@" + message.Author.ID + "> " + message.Author.String()
	send.Embed.Fields[0].Inline = false
	send.Embed.Fields[1].Name = "\U0001f538 UserID:"
	send.Embed.Fields[1].Value = message.Author.ID
	send.Embed.Fields[1].Inline = false
	send.Embed.Fields[2].Name = "\U0001f539 Server Name:"
	send.Embed.Fields[2].Value = state.Invite.Guild.Name
	send.Embed.Fields[2].Inline = false
	send.Embed.Fields[3].Name = "\U0001f538 Role on Server:"
	send.Embed.Fields[3].Value = state.Role
	send.Embed.Fields[3].Inline = false
	send.Embed.Fields[4].Name = "\U0001f539 Server Invite:"
	send.Embed.Fields[4].Value = "https://discord.gg/" + state.Invite.Code
	send.Embed.Fields[4].Inline = false
	send.Embed.Fields[5].Name = "\U0001f538 Server Description:"
	send.Embed.Fields[5].Value = state.Desc
	send.Embed.Fields[5].Inline = false
	send.Embed.Fields[6].Name = "\U0001f539 Number of Members:"
	send.Embed.Fields[6].Value = state.Count
	//      strconv.FormatInt (int64 (state.Invite.Guild.MemberCount), 10)
	send.Embed.Fields[6].Inline = false
	send.Embed.Fields[7].Name = "\U0001f538 Reason to Partner:"
	send.Embed.Fields[7].Value = state.Why
	send.Embed.Fields[7].Inline = false
	send.Embed.Fields[8].Name = "\U0001f539 Additional Comment/Information:"
	send.Embed.Fields[8].Value = message.Content
	send.Embed.Fields[8].Inline = false
	_, err := session.ChannelMessageSendComplex(RsCID, send)
	if err != nil {
		session.ChannelMessageSend(message.ChannelID,
			"Oh F*ck! Seems there is an issue on our side, we are sorry, "+
				"but you might have to contact our team. DM our Zwei bot "+
				"<@433253290898096139> and we will get back to you soon.\n"+
				"https://i.imgur.com/aSFVNkm.gif")
	}

	session.ChannelMessageSend(message.ChannelID,
		"**Partnership application is done. We will contact you soon.  "+
			"If you haven't heard from us in 5 days, it has been rejected.  "+
			"Be patient.**\nhttps://i.imgur.com/VlfiSi4.gif")
}

func TooLong(
	session *discordgo.Session,
	message *discordgo.Message,
	limit int,
) bool {
	length := len(message.Content)
	switch {
	case length < 1:
		session.ChannelMessageSend(message.ChannelID,
			"That's cool and all, but please enter some words.\n"+
				"https://i.imgur.com/ipUVHEs.gif")
	case length > limit:
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf(
			"We appreciate the enthusiasm, but the answer is too long, "+
				"please keep it under %d characters and try again "+
				"(your answer was %d characters).",
			limit, length))
	default:
		return false
	}
	return true
}

package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Run struct {
	T    string `json:"t"`
	S    string `json:"s"`
	Href string `json:"href"`
}

type Content struct {
	Banner       string              `json:"banner"`
	Detective    string              `json:"detective"`
	NeofetchLogo string              `json:"neofetchLogo"`
	Mascot       string              `json:"mascot"`
	MascotLeft   string              `json:"mascotLeft"`
	MascotRight  string              `json:"mascotRight"`
	MascotBlink  string              `json:"mascotBlink"`
	Boot         [][]Run             `json:"boot"`
	AfterBanner  [][]Run             `json:"afterBanner"`
	Files        []string            `json:"files"`
	Sections     map[string][][]Run  `json:"sections"`
	Help         struct {
		Essentials [][2]string `json:"essentials"`
		Extras     []string    `json:"extras"`
		Footer     string      `json:"footer"`
	} `json:"help"`
}

func envHost() string { return os.Getenv("SITE_HOST") }

var fileToCmd = map[string]string{
	"about.txt": "about", "experience.txt": "work", "skills.txt": "skills",
	"education.txt": "edu", "contact.txt": "contact", "interests.txt": "interests",
	"resume.pdf": "resume",
}

func txt(t, s string) string { return styleFor[s].Render(t) }

func sectionLines(name string) []string {
	var out []string
	for _, runs := range content.Sections[name] {
		out = append(out, renderLine(runs))
	}
	return out
}

func artLines(art string) []string {
	var out []string
	for _, l := range strings.Split(art, "\n") {
		out = append(out, stArtDark.Render(l))
	}
	return out
}

func helpLines() []string {
	out := []string{stDim.Render("the essentials:"), ""}
	for _, e := range content.Help.Essentials {
		out = append(out, stKey.Render(fmt.Sprintf("  %-11s", e[0]))+stDim.Render(e[1]))
	}
	out = append(out, "", stDim.Render("for the terminal romantics:"), "")
	for _, x := range content.Help.Extras {
		out = append(out, stKey.Render("  "+x))
	}
	out = append(out, "", stFaint.Render(content.Help.Footer))
	return out
}

func neofetchLines() []string {
	specs := [][2]string{
		{"os", "lukeOS noir"}, {"host", "norfolk, va"}, {"kernel", "spring-boot-21"},
		{"uptime", "1.3 yrs (professional)"}, {"shell", "angular-cli"}, {"wm", "devsecops"},
		{"pkgs", "java python ts c++"}, {"clearance", "ts/sci fsp"},
		{"theme", "bone-on-black"}, {"icons", "none. we use text."},
	}
	var art []string
	mw := 0
	for _, l := range strings.Split(content.NeofetchLogo, "\n") {
		if len(l) > 0 {
			art = append(art, l)
		}
		if len([]rune(l)) > mw {
			mw = len([]rune(l))
		}
	}
	head := "guest@lcotnoir"
	var out []string
	n := len(specs) + 2
	if len(art) > n {
		n = len(art)
	}
	for i := 0; i < n; i++ {
		a := ""
		if i < len(art) {
			a = art[i]
		}
		left := stArtDark.Render(fmt.Sprintf("%-*s", mw+4, a))
		var right string
		switch {
		case i == 0:
			right = stAccent.Render(head)
		case i == 1:
			right = stFaint.Render(strings.Repeat("─", len(head)))
		case i-2 < len(specs):
			s := specs[i-2]
			right = stBody.Render(fmt.Sprintf("%-10s ", s[0])) + stDim.Render(s[1])
		}
		out = append(out, left+right)
	}
	return out
}

func manLines(args []string) []string {
	if len(args) == 0 || args[0] != "luke" {
		q := ""
		if len(args) > 0 {
			q = " " + args[0]
		}
		return []string{stRed.Render("no manual entry for" + q)}
	}
	rows := []struct{ t, s string }{
		{"LUKE(1)                    User Commands                   LUKE(1)", "bright"},
		{"", "body"},
		{"NAME", "body"},
		{"       luke - full-stack engineer, secure systems", "body"},
		{"", "body"},
		{"SYNOPSIS", "body"},
		{"       luke [--angular] [--spring] [--cleared] <mission>", "body"},
		{"", "body"},
		{"DESCRIPTION", "body"},
		{"       ships production software for places where shipping", "body"},
		{"       matters. migrates legacy frontends. runs containers", "body"},
		{"       into classified environs. documents everything.", "body"},
		{"", "body"},
		{"BUGS", "body"},
		{"       occasionally says \"one more song\" at 2am.", "body"},
		{"", "body"},
		{"SEE ALSO", "body"},
		{"       contact(1), hire(1), resume(1)", "body"},
		{"", "body"},
		{"LUKE(1)                    October 2026                    LUKE(1)", "bright"},
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, txt(r.t, r.s))
	}
	return out
}

func htopLines() []string {
	return []string{
		stDim.Render("  PID  CPU%  MEM%  PROCESS"),
		"    1   4.2   12   spring-boot serve",
		"  105  31.0   44   ng migrate --from react",
		"  208   8.8    9   kubectl rollout watch",
		"  313   0.2   61   hollow-knight --silksong --wait",
		stRed.Render("  404   0.0    0   motivation.service [failed, restarted]"),
		"  700  12.1    7   guitar-amp-sim",
		"  911   2.4    3   zillow --scroll --endless",
		"",
		stFaint.Render("load average: reasonable. mostly."),
	}
}

func dispatch(raw string) (lines []string, quit bool) {
	parts := strings.Fields(strings.ToLower(raw))
	if len(parts) == 0 {
		return nil, false
	}
	cmd, args := parts[0], parts[1:]

	switch cmd {
	case "help":
		return helpLines(), false
	case "about", "whoami", "whois":
		out := append(artLines(content.Detective), "")
		return append(out, sectionLines("about")...), false
	case "work", "experience":
		return sectionLines("work"), false
	case "skills":
		return sectionLines("skills"), false
	case "edu", "education":
		out := sectionLines("edu")
		sk := sectionLines("skills")
		return append(append(out, ""), sk[len(sk)-2:]...), false
	case "interests", "hobbies":
		return sectionLines("interests"), false
	case "contact":
		return sectionLines("contact"), false
	case "hire":
		return sectionLines("hire"), false
	case "ls":
		return []string{
			stBright.Render(strings.Join(content.Files, "   ")),
			stFaint.Render("cat <file> works too."),
		}, false
	case "cat":
		if len(args) == 0 {
			return []string{stRed.Render("cat: missing file operand")}, false
		}
		f := args[0]
		if f == "secrets.txt" {
			return []string{stRed.Render("cat: secrets.txt: permission denied")}, false
		}
		target, ok := fileToCmd[f]
		if !ok {
			return []string{stRed.Render(fmt.Sprintf("cat: %s: no such file", f))}, false
		}
		return dispatch(target)
	case "resume":
		return []string{
			stDim.Render("resume.pdf lives on the web build:"),
			stAccent.Render("  https://" + host() + "/assets/resume.pdf"),
		}, false
	case "neofetch":
		return neofetchLines(), false
	case "htop":
		return htopLines(), false
	case "man":
		return manLines(args), false
	case "clearance":
		return []string{
			stDim.Render("> querying adjudication ..."),
			stAccent.Render("> status: ACTIVE TS/SCI — full scope polygraph"),
			"> access granted. you may read this terminal.",
		}, false
	case "sudo":
		j := strings.Join(args, " ")
		if j == "hire luke" || j == "hire luke cotnoir" || j == "hire" {
			return []string{stAccent.Render("permission granted. finally, a good use of sudo.")}, false
		}
		return []string{
			stRed.Render("guest is not in the sudoers file."),
			stDim.Render("this incident will be reported."),
		}, false
	case "ssh":
		target := "guest@" + host()
		if len(args) > 0 {
			target = args[0]
		}
		return []string{
			stDim.Render("ssh " + target),
			stFaint.Render("you are already inside it. ssh-ception."),
		}, false
	case "exit", "logout", "quit":
		return []string{stDim.Render("logout — the fog rolls back in.")}, true
	case "echo":
		return []string{strings.Join(args, " ")}, false
	case "pwd":
		return []string{"/home/guest — you never left"}, false
	case "date":
		return []string{stDim.Render(time.Now().String())}, false
	case "uname":
		return []string{"lukeOS noir 2.6.1 tty1 x86_64"}, false
	case "id":
		return []string{"uid=1000(guest) gid=1000(visitor) groups=cool_people"}, false
	case "ping":
		return []string{"pong. latency: zero."}, false
	case "clear", "cls":
		return nil, false // handled via ctrl+l semantics; keep simple: wipe
	case "vim", "emacs", "nano":
		return []string{stDim.Render("no editor installed. you get a prompt and you like it.")}, false
	case "coffee":
		return []string{
			stAccent.Render("        ( ("), stAccent.Render("         ) )"),
			stAccent.Render("      ........"), stAccent.Render("      |      |]"),
			stAccent.Render("      \\      /"), stAccent.Render("       \\____/"),
			stDim.Render("  black, obviously."),
		}, false
	case "theme":
		if len(args) > 0 && (args[0] == "green" || args[0] == "matrix") {
			return []string{stRed.Render("green phosphor is overdone. staying noir.")}, false
		}
		return []string{stDim.Render("one theme. it is called noir.")}, false
	case "rm":
		return []string{stRed.Render("rm: refused. this is a portfolio, not a prod box.")}, false
	case "peek", "mascot":
		out := append(artLines(content.Mascot), stDim.Render("he sees you."))
		return out, false
	case "games":
		return []string{stDim.Render("overwatch · wow · hollow knight.")}, false
	case "music":
		return []string{stDim.Render("ex-producer. guitar in the corner. always hunting the next record.")}, false
	default:
		if strings.HasPrefix(cmd, "./") || strings.HasPrefix(cmd, "/") {
			return []string{stRed.Render(cmd + ": permission denied")}, false
		}
		return []string{
			stRed.Render("command not found: " + cmd),
			stFaint.Render("type 'help'."),
		}, false
	}
}

var cmdNames = []string{
	"help", "about", "whoami", "whois", "work", "experience", "skills", "edu",
	"education", "interests", "hobbies", "contact", "hire", "ls", "cat", "resume",
	"neofetch", "htop", "man", "clearance", "sudo", "ssh", "exit", "logout",
	"quit", "echo", "pwd", "date", "uname", "id", "ping", "clear", "cls",
	"vim", "emacs", "nano", "coffee", "theme", "rm", "games", "music", "peek",
}

func complete(v string) string {
	parts := strings.Fields(v + " ")
	if len(parts) > 1 && parts[0] == "cat" {
		frag := parts[1]
		var m []string
		for _, f := range content.Files {
			if strings.HasPrefix(f, frag) {
				m = append(m, f)
			}
		}
		if len(m) == 1 {
			return "cat " + m[0]
		}
		return ""
	}
	frag := parts[0]
	var m []string
	for _, c := range cmdNames {
		if strings.HasPrefix(c, frag) {
			m = append(m, c)
		}
	}
	if len(m) == 1 && m[0] != frag {
		return m[0]
	}
	return ""
}

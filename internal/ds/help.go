package ds

// HelpBlock is the in-app /help body with ASCII framing.
func HelpBlock() string {
	return Build(StandardInner, "termcord commands", []string{
		"  /join #name      switch channel",
		"  /threads         list active threads",
		"  /thread name     open a thread",
		"  /parent          leave thread",
		"  /history         load older messages",
		"  /reply           reply to selected message",
		"  /react emoji     toggle reaction",
		"  /search query    search local cache",
		"  /plugins         list loaded plugins",
		"  /cordy           scritch cordy (same as ctrl+y)",
		"  /clear           clear screen (cache kept)",
		"  /quit            exit",
		"  keys",
		"  ctrl+g spotlight · ctrl+m @mentions · ctrl+f focus",
		"  tab @user #channel · ctrl+y scritch cordy · ctrl+b sidebar · ctrl+t threads",
		"  ctrl+r react · ctrl+p/n channels · ctrl+↑/↓ pick message",
		"  pgup/pgdn scroll · ctrl+l jump down · ctrl+c quit",
	})
}

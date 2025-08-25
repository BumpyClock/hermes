package dom

import "regexp"

// Spacer images to be removed
var SPACER_RE = regexp.MustCompile(`(?i)transparent|spacer|blank`)

// The class we will use to mark elements we want to keep
// but would normally remove
const KEEP_CLASS = "hermes-parser-keep"

var KEEP_SELECTORS = []string{
	`iframe[src^="https://www.youtube.com"]`,
	`iframe[src^="https://www.youtube-nocookie.com"]`,
	`iframe[src^="http://www.youtube.com"]`,
	`iframe[src^="https://player.vimeo"]`,
	`iframe[src^="http://player.vimeo"]`,
	`iframe[src^="https://www.redditmedia.com"]`,
}

// A list of tags to strip from the output if we encounter them.
var STRIP_OUTPUT_TAGS = []string{
	"title",
	"script",
	"noscript",
	"link",
	"style",
	"hr",
	"embed",
	"iframe",
	"object",
}

// cleanAttributes
var REMOVE_ATTRS = []string{"style", "align"}

func GetRemoveAttrSelectors() []string {
	selectors := make([]string, len(REMOVE_ATTRS))
	for i, attr := range REMOVE_ATTRS {
		selectors[i] = "[" + attr + "]"
	}
	return selectors
}

var REMOVE_ATTR_LIST = "style,align"

var WHITELIST_ATTRS = []string{
	"src",
	"srcset",
	"sizes",
	"type",
	"href",
	"class",
	"id",
	"alt",
	"xlink:href",
	"width",
	"height",
}

var WHITELIST_ATTRS_RE = regexp.MustCompile(`(?i)^(src|srcset|sizes|type|href|class|id|alt|xlink:href|width|height)$`)

// removeEmpty
var REMOVE_EMPTY_TAGS = []string{"p"}

func GetRemoveEmptySelectors() []string {
	selectors := make([]string, len(REMOVE_EMPTY_TAGS))
	for i, tag := range REMOVE_EMPTY_TAGS {
		selectors[i] = tag + ":empty"
	}
	return selectors
}

var REMOVE_EMPTY_SELECTORS = "p:empty"

// cleanTags
var CLEAN_CONDITIONALLY_TAGS = []string{
	"ul",
	"ol",
	"table",
	"div",
	"button",
	"form",
}

var CLEAN_CONDITIONALLY_TAGS_LIST = "ul,ol,table,div,button,form"

// cleanHeaders
var HEADER_TAGS = []string{"h2", "h3", "h4", "h5", "h6"}
var HEADER_TAG_LIST = "h2,h3,h4,h5,h6"

// CONTENT FETCHING CONSTANTS

// A list of strings that can be considered unlikely candidates when
// extracting content from a resource. These strings are joined together
// and then tested for existence using regex, so may contain simple,
// non-pipe style regular expression queries if necessary.
var UNLIKELY_CANDIDATES_BLACKLIST = []string{
	"ad-break",
	"adbox",
	"advert",
	"addthis",
	"agegate",
	"aux",
	"blogger-labels",
	"combx",
	"comment",
	"conversation",
	"disqus",
	"entry-unrelated",
	"extra",
	"foot",
	// "form", // This is too generic, has too many false positives
	"header",
	"hidden",
	"loader",
	"login", // Note: This can hit 'blogindex'.
	"menu",
	"meta",
	"nav",
	"outbrain",
	"pager",
	"pagination",
	"predicta", // readwriteweb inline ad box
	"presence_control_external", // lifehacker.com container full of false positives
	"popup",
	"printfriendly",
	"related",
	"remove",
	"remark",
	"rss",
	"share",
	"shoutbox",
	"sidebar",
	"sociable",
	"sponsor",
	"taboola",
	"tools",
}

// A list of strings that can be considered LIKELY candidates when
// extracting content from a resource. Essentially, the inverse of the
// blacklist above - if something matches both blacklist and whitelist,
// it is kept. This is useful, for example, if something has a className
// of "rss-content entry-content". It matched 'rss', so it would normally
// be removed, however, it's also the entry content, so it should be left
// alone.
//
// These strings are joined together and then tested for existence using
// regex, so may contain simple, non-pipe style regular expression queries
// if necessary.
var UNLIKELY_CANDIDATES_WHITELIST = []string{
	"and",
	"article",
	"body",
	"blogindex",
	"column",
	"content",
	"entry-content-asset",
	"format", // misuse of form
	"hfeed",
	"hentry",
	"hatom",
	"main",
	"page",
	"posts",
	"shadow",
}

// A list of tags which, if found inside, should cause a <div /> to NOT
// be turned into a paragraph tag. Shallow div tags without these elements
// should be turned into <p /> tags.
var DIV_TO_P_BLOCK_TAGS = []string{
	"a",
	"blockquote",
	"dl",
	"div",
	"img",
	"p",
	"pre",
	"table",
}

var DIV_TO_P_BLOCK_TAGS_LIST = "a,blockquote,dl,div,img,p,pre,table"

// A list of tags that should be ignored when trying to find the top candidate
// for a document.
var NON_TOP_CANDIDATE_TAGS = []string{
	"br",
	"b",
	"i",
	"label",
	"hr",
	"area",
	"base",
	"basefont",
	"input",
	"img",
	"link",
	"meta",
}

var NON_TOP_CANDIDATE_TAGS_RE = regexp.MustCompile(`(?i)^(br|b|i|label|hr|area|base|basefont|input|img|link|meta)$`)

// A list of selectors that specify, very clearly, either hNews or other
// very content-specific style content, like Blogger templates.
// More examples here: http://microformats.org/wiki/blog-post-formats
var HNEWS_CONTENT_SELECTORS = [][]string{
	{".hentry", ".entry-content"},
	{"entry", ".entry-content"},
	{".entry", ".entry_content"},
	{".post", ".postbody"},
	{".post", ".post_body"},
	{".post", ".post-body"},
}

var PHOTO_HINTS = []string{"figure", "photo", "image", "caption"}
var PHOTO_HINTS_RE = regexp.MustCompile(`(?i)figure|photo|image|caption`)

// A list of strings that denote a positive scoring for this content as being
// an article container. Checked against className and id.
//
// TODO: Perhaps have these scale based on their odds of being quality?
var POSITIVE_SCORE_HINTS = []string{
	"article",
	"articlecontent",
	"instapaper_body",
	"blog",
	"body",
	"content",
	"entry-content-asset",
	"entry",
	"hentry",
	"main",
	"Normal",
	"page",
	"pagination",
	"permalink",
	"post",
	"story",
	"text",
	"[-_]copy", // usatoday
	`\\Bcopy`,
}

// The above list, joined into a matching regular expression
var POSITIVE_SCORE_RE = regexp.MustCompile(`(?i)article|articlecontent|instapaper_body|blog|body|content|entry-content-asset|entry|hentry|main|Normal|page|pagination|permalink|post|story|text|[-_]copy|\\Bcopy`)

// Readability publisher-specific guidelines
var READABILITY_ASSET = regexp.MustCompile(`(?i)entry-content-asset`)

// A list of strings that denote a negative scoring for this content as being
// an article container. Checked against className and id.
//
// TODO: Perhaps have these scale based on their odds of being quality?
var NEGATIVE_SCORE_HINTS = []string{
	"adbox",
	"advert",
	"author",
	"bio",
	"bookmark",
	"bottom",
	"byline",
	"clear",
	"com-",
	"combx",
	"comment",
	`comment\\B`,
	"contact",
	"copy",
	"credit",
	"crumb",
	"date",
	"deck",
	"excerpt",
	"featured", // tnr.com has a featured_content which throws us off
	"foot",
	"footer",
	"footnote",
	"graf",
	"head",
	"info",
	"infotext", // newscientist.com copyright
	"instapaper_ignore",
	"jump",
	"linebreak",
	"link",
	"masthead",
	"media",
	"meta",
	"modal",
	"outbrain", // slate.com junk
	"promo",
	"pr_", // autoblog - press release
	"related",
	"respond",
	"roundcontent", // lifehacker restricted content warning
	"scroll",
	"secondary",
	"share",
	"shopping",
	"shoutbox",
	"side",
	"sidebar",
	"sponsor",
	"stamp",
	"sub",
	"summary",
	"tags",
	"tools",
	"widget",
}

// The above list, joined into a matching regular expression
var NEGATIVE_SCORE_RE = regexp.MustCompile(`(?i)adbox|advert|author|bio|bookmark|bottom|byline|clear|com-|combx|comment|comment\\B|contact|copy|credit|crumb|date|deck|excerpt|featured|foot|footer|footnote|graf|head|info|infotext|instapaper_ignore|jump|linebreak|link|masthead|media|meta|modal|outbrain|promo|pr_|related|respond|roundcontent|scroll|secondary|share|shopping|shoutbox|side|sidebar|sponsor|stamp|sub|summary|tags|tools|widget`)

// Additional scoring constants

// XPath to try to determine if a page is wordpress. Not always successful.
const IS_WP_SELECTOR = `meta[name="generator"][value^="WordPress"]`

// Match a digit. Pretty clear.
var DIGIT_RE = regexp.MustCompile(`[0-9]`)

// A list of words that, if found in link text or URLs, likely mean that
// this link is not a next page link.
var EXTRANEOUS_LINK_HINTS = []string{
	"print",
	"archive",
	"comment",
	"discuss",
	"e-mail",
	"email",
	"share",
	"reply",
	"all",
	"login",
	"sign",
	"single",
	"adx",
	"entry-unrelated",
}

var EXTRANEOUS_LINK_HINTS_RE = regexp.MustCompile(`(?i)print|archive|comment|discuss|e-mail|email|share|reply|all|login|sign|single|adx|entry-unrelated`)

// Match any phrase that looks like it could be page, or paging, or pagination
var PAGE_RE = regexp.MustCompile(`(?i)pag(e|ing|inat)`)

// Match any link text/classname/id that looks like it could mean the next
// page. Things like: next, continue, >, >>, » but not >|, »| as those can
// mean last page.
var NEXT_LINK_TEXT_RE = regexp.MustCompile(`(?i)(next|weiter|continue|>([^|]|$)|»([^|]|$))`)

// Match any link text/classname/id that looks like it is an end link: things
// like "first", "last", "end", etc.
var CAP_LINK_TEXT_RE = regexp.MustCompile(`(?i)(first|last|end)`)

// Match any link text/classname/id that looks like it means the previous
// page.
var PREV_LINK_TEXT_RE = regexp.MustCompile(`(?i)(prev|earl|old|new|<|«)`)

// Match 2 or more consecutive <br> tags
var BR_TAGS_RE = regexp.MustCompile(`(?i)(<br[^>]*>[ \n\r\t]*){2,}`)

// Match 1 BR tag.
var BR_TAG_RE = regexp.MustCompile(`(?i)<br[^>]*>`)

// A list of all of the block level tags known in HTML5 and below. Taken from
// http://bit.ly/qneNIT
var BLOCK_LEVEL_TAGS = []string{
	"article",
	"aside",
	"blockquote",
	"body",
	"br",
	"button",
	"canvas",
	"caption",
	"col",
	"colgroup",
	"dd",
	"div",
	"dl",
	"dt",
	"embed",
	"fieldset",
	"figcaption",
	"figure",
	"footer",
	"form",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"header",
	"hgroup",
	"hr",
	"li",
	"map",
	"object",
	"ol",
	"output",
	"p",
	"pre",
	"progress",
	"section",
	"table",
	"tbody",
	"textarea",
	"tfoot",
	"th",
	"thead",
	"tr",
	"ul",
	"video",
}

var BLOCK_LEVEL_TAGS_RE = regexp.MustCompile(`(?i)^(article|aside|blockquote|body|br|button|canvas|caption|col|colgroup|dd|div|dl|dt|embed|fieldset|figcaption|figure|footer|form|h1|h2|h3|h4|h5|h6|header|hgroup|hr|li|map|object|ol|output|p|pre|progress|section|table|tbody|textarea|tfoot|th|thead|tr|ul|video)$`)

// The removal is implemented as a blacklist and whitelist, this test finds
// blacklisted elements that aren't whitelisted. We do this all in one
// expression-both because it's only one pass, and because this skips the
// serialization for whitelisted nodes.
var candidatesBlacklist = "ad-break|ad-banner|adbox|advert|addthis|agegate|aux|blogger-labels|combx|comment|conversation|disqus|entry-unrelated|extra|foot|header|hidden|loader|login|menu|meta|nav|outbrain|pager|pagination|predicta|presence_control_external|popup|printfriendly|related|remove|remark|rss|share|shoutbox|sidebar|sociable|sponsor|taboola|tools"
var CANDIDATES_BLACKLIST = regexp.MustCompile(`(?i)(` + candidatesBlacklist + `)`)

var candidatesWhitelist = "and|article|body|blogindex|column|content|entry-content-asset|format|hfeed|hentry|hatom|main|page|posts|shadow"
var CANDIDATES_WHITELIST = regexp.MustCompile(`(?i)(` + candidatesWhitelist + `)`)

var UNLIKELY_RE = regexp.MustCompile(`(?i)!(` + candidatesWhitelist + `)|(` + candidatesBlacklist + `)`)

var PARAGRAPH_SCORE_TAGS = regexp.MustCompile(`(?i)^(p|li|span|pre)$`)
var CHILD_CONTENT_TAGS = regexp.MustCompile(`(?i)^(td|blockquote|ol|ul|dl)$`)
var BAD_TAGS = regexp.MustCompile(`(?i)^(address|form)$`)

var HTML_OR_BODY_RE = regexp.MustCompile(`(?i)^(html|body)$`)
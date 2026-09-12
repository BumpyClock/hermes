package dom

import "regexp"

// Spacer images to be removed.
var SPACER_RE = regexp.MustCompile(`(?i)transparent|spacer|blank`)

// The class we will use to mark elements we want to keep
// but would normally remove.
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

// cleanAttributes.
var REMOVE_ATTRS = []string{"style", "align"}

var WHITELIST_ATTRS_RE = regexp.MustCompile(`(?i)^(src|srcset|sizes|type|href|class|id|alt|xlink:href|width|height)$`)

// removeEmpty.
var REMOVE_EMPTY_TAGS = []string{"p"}

var REMOVE_EMPTY_SELECTORS = "p:empty"

// cleanTags.
var CLEAN_CONDITIONALLY_TAGS = []string{
	"ul",
	"ol",
	"table",
	"div",
	"button",
	"form",
}

var CLEAN_CONDITIONALLY_TAGS_LIST = "ul,ol,table,div,button,form"

// cleanHeaders.
var HEADER_TAG_LIST = "h2,h3,h4,h5,h6"

// A list of tags which, if found inside, should cause a <div /> to NOT
// be turned into a paragraph tag. Shallow div tags without these elements
// should be turned into <p /> tags.
var DIV_TO_P_BLOCK_TAGS_LIST = "a,blockquote,dl,div,img,p,pre,table"

// A list of tags that should be ignored when trying to find the top candidate
// for a document.
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

var PHOTO_HINTS_RE = regexp.MustCompile(`(?i)figure|photo|image|caption`)

// Positive article-content indicators, checked against className and id.
var POSITIVE_SCORE_RE = regexp.MustCompile(`(?i)article|articlecontent|instapaper_body|blog|body|content|entry-content-asset|entry|hentry|main|Normal|page|pagination|permalink|post|story|text|[-_]copy|\\Bcopy`)

// Readability publisher-specific guidelines.
var READABILITY_ASSET = regexp.MustCompile(`(?i)entry-content-asset`)

// Negative article-content indicators, checked against className and id.
var NEGATIVE_SCORE_RE = regexp.MustCompile(`(?i)adbox|advert|author|bio|bookmark|bottom|byline|clear|com-|combx|comment|comment\\B|contact|copy|credit|crumb|date|deck|excerpt|featured|foot|footer|footnote|graf|head|info|infotext|instapaper_ignore|jump|linebreak|link|masthead|media|meta|modal|outbrain|promo|pr_|related|respond|roundcontent|scroll|secondary|share|shopping|shoutbox|side|sidebar|sponsor|stamp|sub|summary|tags|tools|widget`)

// A list of all of the block level tags known in HTML5 and below. Taken from
// http://bit.ly/qneNIT
var BLOCK_LEVEL_TAGS_RE = regexp.MustCompile(`(?i)^(article|aside|blockquote|body|br|button|canvas|caption|col|colgroup|dd|div|dl|dt|embed|fieldset|figcaption|figure|footer|form|h1|h2|h3|h4|h5|h6|header|hgroup|hr|li|map|object|ol|output|p|pre|progress|section|table|tbody|textarea|tfoot|th|thead|tr|ul|video)$`)

// The removal is implemented as a blacklist and whitelist, this test finds
// blacklisted elements that aren't whitelisted. We do this all in one
// expression-both because it's only one pass, and because this skips the
// serialization for whitelisted nodes.
var candidatesBlacklist = "ad-break|ad-banner|adbox|advert|addthis|agegate|aux|blogger-labels|combx|comment|conversation|disqus|entry-unrelated|extra|foot|header|hidden|loader|login|menu|meta|nav|outbrain|pager|pagination|predicta|presence_control_external|popup|printfriendly|related|remove|remark|rss|share|shoutbox|sidebar|sociable|sponsor|taboola|tools"
var CANDIDATES_BLACKLIST = regexp.MustCompile(`(?i)(` + candidatesBlacklist + `)`)

var candidatesWhitelist = "and|article|body|blogindex|column|content|entry-content-asset|format|hfeed|hentry|hatom|main|page|posts|shadow"
var CANDIDATES_WHITELIST = regexp.MustCompile(`(?i)(` + candidatesWhitelist + `)`)

var PARAGRAPH_SCORE_TAGS = regexp.MustCompile(`(?i)^(p|li|span|pre)$`)
var CHILD_CONTENT_TAGS = regexp.MustCompile(`(?i)^(td|blockquote|ol|ul|dl)$`)
var BAD_TAGS = regexp.MustCompile(`(?i)^(address|form)$`)

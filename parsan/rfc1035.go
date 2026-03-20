package parsan

// suggestConstRuneUnless returns a SuggestionFunc that suggests replacing the first
// character of the input with a constant rune, unless that character matches the
// exception.
//
// The returned function behaves as follows:
//   - Returns nil if the input is empty
//   - Returns nil if the first character equals the exception rune
//   - Otherwise, returns a single suggestion that replaces the first character
//     with the suggested rune
//
// This is useful for replacing invalid characters with a default value while
// allowing certain characters (like separators) to be handled by other rules.
func suggestConstRuneUnless(suggested, exception rune) SuggestionFunc {
	return func(in string) []*parseResult {
		if len(in) == 0 {
			return nil
		}
		first := rune(in[0])
		if first == exception {
			return nil
		}
		remaining := ""
		if len(in) > 1 {
			remaining = in[1:]
		}
		return []*parseResult{
			{
				consumedText: string(suggested),
				rest:         remaining,
			},
		}
	}
}

// suggestConstStringsIf returns a SuggestionFunc that suggests multiple replacement
// strings when the first character of the input matches a specific character.
//
// The returned function behaves as follows:
//   - Returns nil if the input is empty
//   - Returns nil if the first character does not equal the expected rune
//   - Otherwise, returns one suggestion for each string in the suggested slice,
//     each consuming the first character and replacing it with the suggested string
//
// This enables the parser to explore multiple sanitization alternatives. For example,
// an '@' character might be replaced with either "-at-" or "-", allowing both
// possibilities to be evaluated.
func suggestConstStringsIf(suggested []string, expected rune) SuggestionFunc {
	return func(in string) []*parseResult {
		if len(in) == 0 {
			return nil
		}
		first := rune(in[0])
		if first != expected {
			return nil
		}
		remaining := ""
		if len(in) > 1 {
			remaining = in[1:]
		}
		checked := make([]*parseResult, len(suggested))
		for i, s := range suggested {
			checked[i] = &parseResult{
				consumedText: s,
				rest:         remaining,
			}
		}
		return checked
	}
}

// RFC1035Label returns a Rule that validates and sanitizes DNS labels according
// to RFC 1035 Section 2.3.1.
//
// The grammar for a label is:
//
//	<label> ::= <letter> [ [ <ldh-str> ] <let-dig> ]
//
// Valid labels must:
//   - Start with a letter (a-z, A-Z)
//   - End with a letter or digit (if longer than one character)
//   - Contain only letters, digits, or hyphens in between
//   - Be at most 63 characters long
//
// Sanitization strategies:
//   - Invalid first character: prepended or replaced with 'x'
//   - Invalid last character: replaced with 'x'
//   - Invalid middle characters: handled by the provided suggestFn
func RFC1035Label(suggestFn SuggestionFunc) Rule {
	return GenerateRFC1035Label(suggestFn)
}

// RFC1035LabelRelaxed returns a Rule similar to RFC1035Label but allows labels
// to start with a digit in addition to letters. This relaxation is common in
// practice, as many systems accept labels beginning with digits despite the
// strict RFC 1035 grammar.
func RFC1035LabelRelaxed(suggestFn SuggestionFunc) Rule {
	return GenerateRFC1035Label(suggestFn).MayStartWithDigit(true)
}

// RFC1035LowerLabel returns a Rule identical to RFC1035Label but restricts
// letters to lowercase only (a-z). This is useful when case-normalized labels
// are required, enabling case-insensitive comparisons via exact string matching.
func RFC1035LowerLabel(suggestFn SuggestionFunc) Rule {
	return GenerateRFC1035Label(suggestFn).LowercaseOnly(true)
}

// RFC1035LowerLabelRelaxed returns a Rule that combines the relaxed starting
// character requirement (allowing digits) with lowercase letter enforcement.
// Labels may start with a lowercase letter or digit and contain only lowercase
// letters, digits, and hyphens.
func RFC1035LowerLabelRelaxed(suggestFn SuggestionFunc) Rule {
	return GenerateRFC1035Label(suggestFn).
		MayStartWithDigit(true).
		LowercaseOnly(true)
}

// RFC1035LabelRule is the concrete implementation of a DNS label rule.
// It embeds a generic Rule and adds configuration fields used to build the
// underlying parsing expression.
type RFC1035LabelRule struct {
	Rule
	suggestFn         SuggestionFunc
	lowercase         bool
	mayStartWithDigit bool
	maxLen            int
}

// generate constructs the parsing expression for the label based on the
// current configuration (lowercase enforcement, digit‑start allowance, and
// maximum length) and stores it in the embedded Rule field.
func (rule *RFC1035LabelRule) generate() {
	rule.Rule = Concat(
		rule.firstLetter(),
		Opt(Concat(
			rule.midLetter(),
			rule.lastLetter(),
		)),
	).WithMaxLength(rule.maxLen)
}

// firstLetter creates the rule that matches the first character of a label,
// taking into account whether digits are allowed to start the label and whether
// only lowercase letters are permitted.
func (rule *RFC1035LabelRule) firstLetter() Rule {
	switch {
	case !rule.lowercase && !rule.mayStartWithDigit:
		return Letter(PrependOrReplaceFirstRuneWithStrings("x"))
	case !rule.lowercase && rule.mayStartWithDigit:
		return LetDig(PrependOrReplaceFirstRuneWithStrings("x"))
	case rule.lowercase && !rule.mayStartWithDigit:
		return LowerLetter(PrependOrReplaceFirstRuneWithStrings("x"))
	case rule.lowercase && rule.mayStartWithDigit:
		return LowerLetDig(PrependOrReplaceFirstRuneWithStrings("x"))
	}
	panic("cannot reach this point")
}

// lastLetter creates the rule that matches the final character of a label,
// substituting an invalid character with the placeholder rune “x”.
func (rule *RFC1035LabelRule) lastLetter() Rule {
	if rule.lowercase {
		return LowerLetDig(ReplaceFirstRuneWithStrings("x"))
	}
	return LetDig(ReplaceFirstRuneWithStrings("x"))
}

// midLetter creates the optional rule that validates the middle characters of
// a label (the “ldh‑str” part) and applies the provided suggestion function for
// any invalid characters.
func (rule *RFC1035LabelRule) midLetter() Rule {
	if !rule.lowercase {
		return Opt(LDHStr(rule.suggestFn))
	} else {
		return Opt(LowerLDHStr(rule.suggestFn))
	}
}

// LowercaseOnly configures the rule to accept only lowercase letters.
// It regenerates the underlying expression to reflect the new setting.
func (rule *RFC1035LabelRule) LowercaseOnly(b bool) *RFC1035LabelRule {
	rule.lowercase = b
	rule.generate()
	return rule
}

// MayStartWithDigit configures the rule to allow a digit as the first
// character of a label. It regenerates the underlying expression to
// reflect the new setting.
func (rule *RFC1035LabelRule) MayStartWithDigit(b bool) *RFC1035LabelRule {
	rule.mayStartWithDigit = b
	rule.generate()
	return rule
}

// MaxLength sets the maximum allowed length for the label and rebuilds the
// parsing expression accordingly.
func (rule *RFC1035LabelRule) MaxLength(l int) *RFC1035LabelRule {
	rule.maxLen = l
	rule.generate()
	return rule
}

// GenerateRFC1035Label creates a new RFC1035LabelRule with the
// supplied SuggestionFunc. The rule starts with the strict RFC‑1035
// defaults (lowercase disabled, digit‑start disabled, max length 63)
// and then builds its internal expression.
func GenerateRFC1035Label(suggestFn SuggestionFunc) *RFC1035LabelRule {
	rule := &RFC1035LabelRule{
		suggestFn:         suggestFn,
		lowercase:         false,
		mayStartWithDigit: false,
		maxLen:            63,
	}
	rule.generate()
	return rule
}

// subdomainSuggestFn is a SuggestionFunc that handles invalid characters within
// subdomain labels by combining two strategies:
//
//  1. The '@' character is offered two replacement options: "-at-" or "-"
//  2. All other invalid characters (except '.') are replaced with '-'
//
// The '.' character is explicitly excluded because it serves as the label
// separator in subdomains and must be processed by the subdomain rule's
// concatenation logic rather than being replaced.
var subdomainSuggestFn = MergeSuggestionFuncs(
	suggestConstStringsIf([]string{"-at-", "-"}, '@'),
	suggestConstRuneUnless('-', '.'),
)

// RFC1035Subdomain is a Rule that validates and sanitizes DNS subdomains
// according to RFC 1035 Section 2.3.1.
//
// The grammar for a subdomain is:
//
//	<subdomain> ::= <label> | <subdomain> "." <label>
//
// A subdomain consists of one or more dot-separated labels, where each label
// conforms to RFC1035Label. The total length must not exceed 253 characters.
//
// Invalid characters within labels are sanitized as follows:
//   - '@' is replaced with "-at-" or "-" (both options explored)
//   - Other invalid characters are replaced with '-'
var RFC1035Subdomain = GenerateRFC1035Subdomain()

// RFC1035SubdomainRelaxed is a Rule similar to RFC1035Subdomain but uses
// RFC1035LabelRelaxed for each label, allowing labels to start with digits.
// This accommodates the common practice of using digit-prefixed labels in
// DNS names.
var RFC1035SubdomainRelaxed = GenerateRFC1035Subdomain().LabelMayStartWithDigit(true)

// RFC1035LowerSubdomain is a Rule identical to RFC1035Subdomain but enforces
// lowercase letters throughout all labels. This produces case-normalized
// subdomain strings suitable for systems that perform case-insensitive DNS
// comparisons via exact string matching.
var RFC1035LowerSubdomain = GenerateRFC1035Subdomain().LowercaseOnlyLabel(true)

// RFC1035LowerSubdomainRelaxed is a Rule that combines relaxed starting
// character requirements with lowercase letter enforcement. Labels may
// start with a lowercase letter or digit, and all letters are restricted
// to lowercase. This is the most permissive variant while still enforcing
// case normalization.
var RFC1035LowerSubdomainRelaxed = GenerateRFC1035Subdomain().
	LowercaseOnlyLabel(true).
	LabelMayStartWithDigit(true)

// RFC1035SubdomainRule defines a parser rule for a full DNS subdomain.
// It contains a reference to the label rule used for each component.
type RFC1035SubdomainRule struct {
	Rule
	Label *RFC1035LabelRule
}

// generate builds the recursive parsing expression for a subdomain.
// It uses a unique rule name to allow the rule to refer to itself.
func (r *RFC1035SubdomainRule) generate() {
	name := GenerateUniqueName()
	r.Rule = Named(name,
		Alternative(
			r.Label,
			Concat(
				r.Label,
				Terminal("."),
				Ref(name),
			),
		)).WithMaxLength(253)
}

// LowercaseOnlyLabel configures the inner label rule to enforce
// lowercase letters only, then rebuilds the subdomain expression.
func (r *RFC1035SubdomainRule) LowercaseOnlyLabel(b bool) *RFC1035SubdomainRule {
	r.Label = r.Label.LowercaseOnly(b)
	r.generate()
	return r
}

// LabelMayStartWithDigit configures the inner label rule to allow a
// digit as the first character of each label, then rebuilds the
// subdomain expression.
func (r *RFC1035SubdomainRule) LabelMayStartWithDigit(b bool) *RFC1035SubdomainRule {
	r.Label = r.Label.MayStartWithDigit(b)
	r.generate()
	return r
}

// LabelMaxLength sets a maximum length for each label within the
// subdomain and rebuilds the parsing expression.
func (r *RFC1035SubdomainRule) LabelMaxLength(l int) *RFC1035SubdomainRule {
	r.Label = r.Label.MaxLength(l)
	r.generate()
	return r
}

// GenerateRFC1035Subdomain creates a new subdomain rule using the
// default subdomain suggestion function. The returned rule can be
// further tuned (e.g., relaxed digit start, lowercase enforcement)
// via its methods.
func GenerateRFC1035Subdomain() *RFC1035SubdomainRule {
	r := &RFC1035SubdomainRule{
		Label: GenerateRFC1035Label(subdomainSuggestFn),
	}
	r.generate()
	return r
}

import sanitizeHtml from "sanitize-html";

/**
 * Sanitiser for third-party email HTML.
 *
 * Inline styles are deliberately not allowed. sanitize-html only applies an
 * `allowedStyles` allowlist when `style` is itself an allowed attribute, and it
 * is not in sanitize-html's defaults — so styles are stripped entirely. An
 * elaborate `allowedStyles` block used to sit here and in the inbox page, doing
 * nothing while reading as though inline CSS were being validated. Allowing
 * styles is a product decision, not a formatting detail: it would need the
 * property allowlist reinstated alongside `style` in allowedAttributes.
 *
 * `img` is permitted so newsletters render; sanitize-html's default
 * allowedSchemes keeps `javascript:` and friends out of src and href.
 */
const emailSanitizeOptions: sanitizeHtml.IOptions = {
  allowedTags: sanitizeHtml.defaults.allowedTags.concat(["img"]),
  allowedAttributes: {
    ...sanitizeHtml.defaults.allowedAttributes,
  },
};

/** Sanitises email HTML for rendering. Safe to call with an empty string. */
export function sanitizeEmailHtml(html: string): string {
  if (!html) return "";
  return sanitizeHtml(html, emailSanitizeOptions);
}

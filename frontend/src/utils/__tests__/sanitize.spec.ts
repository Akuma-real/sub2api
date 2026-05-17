import { describe, expect, it } from "vitest";
import { renderMarkdownToSafeHtml, sanitizeHtml, sanitizeSvg } from "../sanitize";

describe("sanitize helpers", () => {
  it("renders markdown help text while stripping unsafe HTML", () => {
    const html = renderMarkdownToSafeHtml(
      "**Important** [docs](https://example.com)<script>alert(1)</script>",
    );

    expect(html).toContain("<strong>Important</strong>");
    expect(html).toContain('<a href="https://example.com">docs</a>');
    expect(html).not.toContain("<script");
  });

  it("sanitizes raw HTML and SVG content with separate profiles", () => {
    expect(sanitizeHtml('<img src=x onerror="alert(1)">')).not.toContain(
      "onerror",
    );
    expect(sanitizeSvg('<svg><script>alert(1)</script><path /></svg>')).not.toContain(
      "<script",
    );
  });
});

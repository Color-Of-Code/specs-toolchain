import type { DetailRow, MountOptions } from "./types";

export function escapeHTML(value: string | number | boolean | null | undefined): string {
  return String(value ?? "").replace(
    /[&<>"']/g,
    (char) => {
      switch (char) {
        case "&":
          return "&amp;";
        case "<":
          return "&lt;";
        case ">":
          return "&gt;";
        case '"':
          return "&quot;";
        case "'":
          return "&#39;";
        default:
          return char;
      }
    },
  );
}

export function detailsIconButton(label: string, path: string | undefined): string {
  if (!path) {
    return "";
  }
  return `<button type="button" class="details-open-button details-icon-button" data-open-path="${escapeHTML(path)}" aria-label="${escapeHTML(label)}" title="${escapeHTML(label)}"><span class="details-visually-hidden">${escapeHTML(label)}</span></button>`;
}

export function detailsRowsMarkup(rows: DetailRow[] | undefined): string {
  const renderedRows = (rows ?? [])
    .filter((row) => row && row.value != null && row.value !== "")
    .map((row) => {
      const linkButton = row.link ? detailsIconButton(`Open ${row.label}`, row.link) : "";
      return `<div><dt><span>${escapeHTML(row.label)}</span>${linkButton}</dt><dd>${escapeHTML(row.value)}</dd></div>`;
    })
    .join("");
  return renderedRows ? `<dl class="details-list">${renderedRows}</dl>` : "";
}

export function detailsMarkup(
  eyebrow: string,
  title: string,
  rows: DetailRow[] | undefined,
  note: string | undefined,
): string {
  const renderedRows = detailsRowsMarkup(rows);
  const noteMarkup = note
    ? `<p class="details-note">${escapeHTML(note)}</p>`
    : "";
  return `<article class="details-panel"><p class="details-eyebrow">${escapeHTML(eyebrow)}</p><h2 class="details-title">${escapeHTML(title)}</h2>${renderedRows}${noteMarkup}</article>`;
}

export function setDetails(options: MountOptions, markup: string): void {
  if (!options.detailsElement) {
    return;
  }
  options.detailsElement.innerHTML = markup;
}

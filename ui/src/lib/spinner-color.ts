// The loading spinner takes one of the section heading colors, picked once
// per page load, so each reload shows a different color of the logo. The
// class names are written out in full so that Tailwind generates them.
const colors = [
  "text-heading-1",
  "text-heading-2",
  "text-heading-3",
  "text-heading-4",
  "text-heading-5",
];

export const spinnerColor = colors[Math.floor(Math.random() * colors.length)];

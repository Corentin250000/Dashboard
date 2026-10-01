export const SIZES = {
  small: { w: 1, h: 1 },
  wide:  { w: 2, h: 1 },
  tall:  { w: 1, h: 2 },
  big:   { w: 2, h: 2 },
} as const;

export type SizeName = keyof typeof SIZES;
import { describe, expect, it } from "vitest";

import { formatBpm } from "./BPMInput";

describe("formatBpm", () => {
  it("should return integer as-is", () => {
    expect(formatBpm("70")).toBe("70");
  });

  it("should accept .5 decimal", () => {
    expect(formatBpm("70.5")).toBe("70.5");
  });

  it("should strip invalid decimal .1", () => {
    expect(formatBpm("70.1")).toBe("70");
  });

  it("should strip invalid decimal .25", () => {
    expect(formatBpm("70.25")).toBe("70");
  });

  it("should normalize .50 to .5", () => {
    expect(formatBpm("70.50")).toBe("70.5");
  });

  it("should strip non-digits", () => {
    expect(formatBpm("abc")).toBe("");
  });

  it("should strip leading zeros", () => {
    expect(formatBpm("070")).toBe("70");
  });

  it("should limit to 3 integer digits", () => {
    expect(formatBpm("1234")).toBe("123");
  });

  it("should allow decimal point for typing .5", () => {
    expect(formatBpm("70.")).toBe("70.");
  });

  it("should handle .5 with max digits", () => {
    expect(formatBpm("300.5")).toBe("300.5");
  });

  it("should strip multiple decimal points", () => {
    expect(formatBpm("70.5.3")).toBe("70.5");
  });

  it("should handle empty string", () => {
    expect(formatBpm("")).toBe("");
  });
});

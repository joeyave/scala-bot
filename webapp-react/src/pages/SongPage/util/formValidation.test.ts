import { describe, expect, it } from "vitest";

import { isBpmValid } from "./formValidation";

describe("isBpmValid", () => {
  it("should return true for empty string", () => {
    expect(isBpmValid("")).toBe(true);
  });

  it("should return true for valid integer", () => {
    expect(isBpmValid("70")).toBe(true);
  });

  it("should return true for valid .5 decimal", () => {
    expect(isBpmValid("70.5")).toBe(true);
  });

  it("should return true for 120.5", () => {
    expect(isBpmValid("120.5")).toBe(true);
  });

  it("should return false for .1 decimal", () => {
    expect(isBpmValid("70.1")).toBe(false);
  });

  it("should return false for .25 decimal", () => {
    expect(isBpmValid("70.25")).toBe(false);
  });

  it("should return false for value below 20", () => {
    expect(isBpmValid("19")).toBe(false);
  });

  it("should return false for value above 300", () => {
    expect(isBpmValid("301")).toBe(false);
  });

  it("should return false for non-numeric", () => {
    expect(isBpmValid("abc")).toBe(false);
  });

  it("should return true for boundary value 20", () => {
    expect(isBpmValid("20")).toBe(true);
  });

  it("should return true for boundary value 300", () => {
    expect(isBpmValid("300")).toBe(true);
  });

  it("should return true for 20.5", () => {
    expect(isBpmValid("20.5")).toBe(true);
  });

  it("should return false for trailing dot", () => {
    expect(isBpmValid("70.")).toBe(false);
  });
});

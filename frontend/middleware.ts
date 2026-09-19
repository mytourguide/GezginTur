import { NextRequest, NextResponse } from "next/server";

// Giris gerektiren sayfalar: yalnizca cookie varligi kontrol edilir (gercek
// yetki kontrolu her istekte API'deki JWT middleware'inde yapilir).
export function middleware(req: NextRequest) {
  const authed = req.cookies.get("auth")?.value === "1";
  if (!authed) {
    const url = req.nextUrl.clone();
    url.pathname = "/giris";
    return NextResponse.redirect(url);
  }
  return NextResponse.next();
}

export const config = { matcher: ["/hesabim/:path*", "/rezervasyon/:path*"] };

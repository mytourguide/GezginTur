export const dynamic = 'force-dynamic';

export default function MesafeliPage() {
  return (
    <main className="max-w-3xl mx-auto px-4 py-16">
      <h1 className="text-3xl font-bold mb-6">Mesafeli Satış Sözleşmesi</h1>
      <article className="prose prose-neutral max-w-none">
        <p><strong>SATICI:</strong> Gezgin Tur Seyahat Acentesi, mytourguide@gmail.com, 0507 654 6712</p>
        <p><strong>ALICI:</strong> Rezervasyon yapan tüketici</p>
        <p>Mesafeli sözleşme kapsamında; tur, otel, uçak bileti gibi seyahat hizmetleri elektronik ortamda satışa sunulur.</p>
        <p>Ödeme öncesinde hizmet içeriği, fiyat, iptal-iade koşulları, sigorta ve ek ücretler açıkça bildirilir. Tüketici onay vermeden sipariş tamamlanmaz.</p>
        <p>İptal ve iade, Tüketicinin Korunması Hakkında Kanun ve TURSAB düzenlemelerine tabidir. Program broşürü ve rezervasyon formu esas alınır.</p>
        <p>Şirket, mücbir sebep ve hava yolu/tur operatörü değişikliklerinde hizmeti uygun alternatifle yerine getirme hakkını saklı tutar.</p>
        <p>Çerez ve iletişim tercihleri konusunda KVKK aydınlatma metni geçerlidir.</p>
      </article>
    </main>
  );
}

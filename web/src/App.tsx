import { FormEvent, useCallback, useEffect, useState } from "react";

type Stock = { sku: string; availableQuantity: number };
type Reservation = {
  id: string;
  sku: string;
  quantity: number;
  status: "pending" | "confirmed" | "released" | "expired";
  expiresAt: string;
  createdAt: string;
};

const products = [
  { sku: "keyboard", name: "Orbit Keyboard", description: "Low-profile mechanical, graphite" },
  { sku: "monitor", name: "Frame 27 Display", description: "4K studio panel, matte finish" },
  { sku: "webcam", name: "Focus Camera", description: "1440p, auto-framing" },
];

const apiBase = import.meta.env.VITE_API_BASE_URL ?? "";

async function readJSON<T>(response: Response): Promise<T> {
  const body = await response.json();
  if (!response.ok) throw new Error(body.error ?? `Request failed with ${response.status}`);
  return body as T;
}

export default function App() {
  const [stock, setStock] = useState<Stock[]>([]);
  const [selectedSKU, setSelectedSKU] = useState(products[0].sku);
  const [quantity, setQuantity] = useState(1);
  const [reservation, setReservation] = useState<Reservation | null>(null);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState<string | null>(null);

  const refreshStock = useCallback(async () => {
    setLoading(true);
    try {
      const responses = await Promise.all(products.map(({ sku }) => fetch(`${apiBase}/api/inventory?sku=${sku}`).then(readJSON<Stock>)));
      setStock(responses);
      setMessage(null);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Inventory is unavailable");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void refreshStock(); }, [refreshStock]);

  async function createReservation(event: FormEvent) {
    event.preventDefault();
    setMessage(null);
    try {
      const created = await fetch(`${apiBase}/api/reservations`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "Idempotency-Key": crypto.randomUUID() },
        body: JSON.stringify({ sku: selectedSKU, quantity, ttlSeconds: 900 }),
      }).then(readJSON<Reservation>);
      setReservation(created);
      await refreshStock();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Could not create the reservation");
    }
  }

  async function transition(action: "confirm" | "release") {
    if (!reservation) return;
    try {
      const updated = await fetch(`${apiBase}/api/reservations/${reservation.id}/${action}`, { method: "POST" }).then(readJSON<Reservation>);
      setReservation(updated);
      setMessage(action === "confirm" ? "Reservation confirmed." : "Stock returned to inventory.");
      await refreshStock();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : `Could not ${action} the reservation`);
    }
  }

  return (
    <main>
      <header className="topbar">
        <a className="brand" href="#top" aria-label="ReserveKit home"><span>RK</span> ReserveKit</a>
        <div className="service-state"><i /> Services connected</div>
      </header>

      <section className="hero" id="top">
        <p className="eyebrow">Inventory operations / live holds</p>
        <h1>Reserve stock.<br /><em>Keep promises.</em></h1>
        <p className="intro">A focused control plane for inspecting availability and exercising the complete reservation lifecycle across an HTTP API and Go gRPC service.</p>
      </section>

      {message && <div className="notice" role="status">{message}</div>}

      <section className="workspace" aria-label="Reservation workspace">
        <div className="inventory-panel">
          <div className="section-heading">
            <div><p className="eyebrow">01 / INVENTORY</p><h2>Available now</h2></div>
            <button className="text-button" type="button" onClick={() => void refreshStock()}>Refresh</button>
          </div>
          <div className="product-grid" aria-busy={loading}>
            {products.map((product) => {
              const available = stock.find((item) => item.sku === product.sku)?.availableQuantity;
              return (
                <button key={product.sku} className={`product-card ${selectedSKU === product.sku ? "selected" : ""}`} type="button" onClick={() => setSelectedSKU(product.sku)}>
                  <span className="product-index">{product.sku.slice(0, 2).toUpperCase()}</span>
                  <strong>{product.name}</strong>
                  <small>{product.description}</small>
                  <span className="quantity">{loading ? "—" : available ?? 0}<small> available</small></span>
                </button>
              );
            })}
          </div>
        </div>

        <aside className="reservation-panel">
          <p className="eyebrow">02 / RESERVATION</p>
          <h2>Create a hold</h2>
          <form onSubmit={createReservation}>
            <label>Product<select value={selectedSKU} onChange={(event) => setSelectedSKU(event.target.value)}>{products.map((product) => <option key={product.sku} value={product.sku}>{product.name}</option>)}</select></label>
            <label>Quantity<input type="number" min="1" max="10" value={quantity} onChange={(event) => setQuantity(Number(event.target.value))} /></label>
            <button className="primary" type="submit">Hold inventory <span>→</span></button>
          </form>

          {reservation ? (
            <div className="receipt">
              <div><span className={`status ${reservation.status}`}>{reservation.status}</span><code>{reservation.id.slice(0, 8)}</code></div>
              <p><strong>{reservation.quantity} × {reservation.sku}</strong><br />Expires {new Date(reservation.expiresAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</p>
              {reservation.status === "pending" && <div className="actions"><button type="button" onClick={() => void transition("confirm")}>Confirm</button><button type="button" className="secondary" onClick={() => void transition("release")}>Release</button></div>}
            </div>
          ) : <p className="empty">Choose an item and create a hold to see the distributed workflow in action.</p>}
        </aside>
      </section>
    </main>
  );
}

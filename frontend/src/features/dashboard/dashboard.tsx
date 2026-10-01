function DashboardPage() {
  return (
    <div className="max-w-sm overflow-hidden rounded-xl bg-white shadow-lg">
      <div className="p-6">
        <h2 className="mb-2 text-xl font-bold text-gray-900">
          Titre de la card
        </h2>
        <p className="mb-4 text-gray-600">
          Une courte description qui explique le contenu de cette card.
        </p>
        <a
          href="#"
          className="inline-block rounded-lg bg-blue-600 px-4 py-2 text-white transition hover:bg-blue-700"
        >
          En savoir plus
        </a>
      </div>
    </div>
  );
}

export default DashboardPage;
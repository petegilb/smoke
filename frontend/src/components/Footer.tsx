function Footer() {
  return (
    <footer className="footer sm:footer-horizontal footer-center bg-base-200 text-base-content/70 p-6 mt-8">
      <aside>
        <p>© {new Date().getFullYear()} smoke</p>
        <p className="text-xs">
          Not affiliated with Valve. Steam and the Steam logo are trademarks of Valve Corporation.
        </p>
      </aside>
    </footer>
  )
}

export default Footer

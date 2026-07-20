<script>
  export let goTo;
</script>

<footer class="border-top mt-5 py-4 bg-light">
  <div class="container">
    <div class="row">
      <div class="col-md-4 mb-3">
        <h5>Polaris</h5>
        <p class="text-muted small">A community-driven social platform.</p>
      </div>
      <div class="col-md-4 mb-3">
        <h5>Links</h5>
        <ul class="list-unstyled">
          <li><button class="btn btn-link p-0 text-decoration-none" on:click={() => goTo({ page: 'about' })}>About</button></li>
          <li><a class="text-decoration-none" href="/docs/" target="_blank">Documentation</a></li>
        </ul>
      </div>
      <div class="col-md-4 mb-3">
        <h5>Legal</h5>
        <ul class="list-unstyled">
          <li><span class="text-muted small">Privacy Policy</span></li>
          <li><span class="text-muted small">Terms of Service</span></li>
        </ul>
      </div>
    </div>
    <div class="text-center text-muted small pt-3 border-top">
      &copy; {new Date().getFullYear()} Polaris — All rights reserved.
    </div>
  </div>
</footer>

<style>
  footer { margin-top: auto; }
  .btn-link { vertical-align: baseline; }
</style>

#include "wrapper.h"

void event(const u32 id, hashcat_ctx_t *hashcat_ctx, const void *buf, const size_t len)
{
  gocat_ctx_t *worker_tuple = (gocat_ctx_t*)hashcat_ctx;
  // call the validator callback if we're in the hash validation mode
  if (worker_tuple->bValidateHashes)
  {
    validatorCallback(id, &worker_tuple->ctx, worker_tuple->gowrapper, (void*)buf, (size_t)len);
  }
  else
  {
    callback(id, &worker_tuple->ctx, worker_tuple->gowrapper, (void*)buf, (size_t)len);
  }
}

void freeargv(int argc, char **argv)
{
  for (int i = 0; i < argc; i++)
  {
    free(argv[i]);
  }
  free(argv);
}

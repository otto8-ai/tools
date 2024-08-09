# Thank you ChatGPT
yq eval '
    .paths |= with_entries(
      .value |= with_entries(
        select(
          ((.key == "get" or .key == "post" or .key == "put" or .key == "delete" or .key == "patch" or .key == "options" or .key == "head" or .key == "trace")
          and (
            .value.operationId == "<operationId>"
          )) or (.key == "description" or .key == "parameters")
        )
      )
    )
    | .paths |= with_entries(
      select(
        .value.get != null or .value.post != null or .value.put != null or .value.delete != null or .value.patch != null or .value.options != null or .value.head != null or .value.trace != null
      )
    )
  '

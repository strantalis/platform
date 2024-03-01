// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: policy/attributes/attributes.proto

// Protobuf Java Version: 3.25.3
package io.opentdf.platform.policy.attributes;

/**
 * Protobuf type {@code policy.attributes.GetAttributesByValueFqnsRequest}
 */
public final class GetAttributesByValueFqnsRequest extends
    com.google.protobuf.GeneratedMessageV3 implements
    // @@protoc_insertion_point(message_implements:policy.attributes.GetAttributesByValueFqnsRequest)
    GetAttributesByValueFqnsRequestOrBuilder {
private static final long serialVersionUID = 0L;
  // Use GetAttributesByValueFqnsRequest.newBuilder() to construct.
  private GetAttributesByValueFqnsRequest(com.google.protobuf.GeneratedMessageV3.Builder<?> builder) {
    super(builder);
  }
  private GetAttributesByValueFqnsRequest() {
    fqns_ =
        com.google.protobuf.LazyStringArrayList.emptyList();
  }

  @java.lang.Override
  @SuppressWarnings({"unused"})
  protected java.lang.Object newInstance(
      UnusedPrivateParameter unused) {
    return new GetAttributesByValueFqnsRequest();
  }

  public static final com.google.protobuf.Descriptors.Descriptor
      getDescriptor() {
    return io.opentdf.platform.policy.attributes.AttributesProto.internal_static_policy_attributes_GetAttributesByValueFqnsRequest_descriptor;
  }

  @java.lang.Override
  protected com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internalGetFieldAccessorTable() {
    return io.opentdf.platform.policy.attributes.AttributesProto.internal_static_policy_attributes_GetAttributesByValueFqnsRequest_fieldAccessorTable
        .ensureFieldAccessorsInitialized(
            io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest.class, io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest.Builder.class);
  }

  public static final int FQNS_FIELD_NUMBER = 1;
  @SuppressWarnings("serial")
  private com.google.protobuf.LazyStringArrayList fqns_ =
      com.google.protobuf.LazyStringArrayList.emptyList();
  /**
   * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
   * @return A list containing the fqns.
   */
  public com.google.protobuf.ProtocolStringList
      getFqnsList() {
    return fqns_;
  }
  /**
   * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
   * @return The count of fqns.
   */
  public int getFqnsCount() {
    return fqns_.size();
  }
  /**
   * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
   * @param index The index of the element to return.
   * @return The fqns at the given index.
   */
  public java.lang.String getFqns(int index) {
    return fqns_.get(index);
  }
  /**
   * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
   * @param index The index of the value to return.
   * @return The bytes of the fqns at the given index.
   */
  public com.google.protobuf.ByteString
      getFqnsBytes(int index) {
    return fqns_.getByteString(index);
  }

  private byte memoizedIsInitialized = -1;
  @java.lang.Override
  public final boolean isInitialized() {
    byte isInitialized = memoizedIsInitialized;
    if (isInitialized == 1) return true;
    if (isInitialized == 0) return false;

    memoizedIsInitialized = 1;
    return true;
  }

  @java.lang.Override
  public void writeTo(com.google.protobuf.CodedOutputStream output)
                      throws java.io.IOException {
    for (int i = 0; i < fqns_.size(); i++) {
      com.google.protobuf.GeneratedMessageV3.writeString(output, 1, fqns_.getRaw(i));
    }
    getUnknownFields().writeTo(output);
  }

  @java.lang.Override
  public int getSerializedSize() {
    int size = memoizedSize;
    if (size != -1) return size;

    size = 0;
    {
      int dataSize = 0;
      for (int i = 0; i < fqns_.size(); i++) {
        dataSize += computeStringSizeNoTag(fqns_.getRaw(i));
      }
      size += dataSize;
      size += 1 * getFqnsList().size();
    }
    size += getUnknownFields().getSerializedSize();
    memoizedSize = size;
    return size;
  }

  @java.lang.Override
  public boolean equals(final java.lang.Object obj) {
    if (obj == this) {
     return true;
    }
    if (!(obj instanceof io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest)) {
      return super.equals(obj);
    }
    io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest other = (io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest) obj;

    if (!getFqnsList()
        .equals(other.getFqnsList())) return false;
    if (!getUnknownFields().equals(other.getUnknownFields())) return false;
    return true;
  }

  @java.lang.Override
  public int hashCode() {
    if (memoizedHashCode != 0) {
      return memoizedHashCode;
    }
    int hash = 41;
    hash = (19 * hash) + getDescriptor().hashCode();
    if (getFqnsCount() > 0) {
      hash = (37 * hash) + FQNS_FIELD_NUMBER;
      hash = (53 * hash) + getFqnsList().hashCode();
    }
    hash = (29 * hash) + getUnknownFields().hashCode();
    memoizedHashCode = hash;
    return hash;
  }

  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      java.nio.ByteBuffer data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      java.nio.ByteBuffer data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      com.google.protobuf.ByteString data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      com.google.protobuf.ByteString data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(byte[] data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      byte[] data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseDelimitedFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseDelimitedWithIOException(PARSER, input);
  }

  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseDelimitedFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseDelimitedWithIOException(PARSER, input, extensionRegistry);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      com.google.protobuf.CodedInputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input);
  }
  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest parseFrom(
      com.google.protobuf.CodedInputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  @java.lang.Override
  public Builder newBuilderForType() { return newBuilder(); }
  public static Builder newBuilder() {
    return DEFAULT_INSTANCE.toBuilder();
  }
  public static Builder newBuilder(io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest prototype) {
    return DEFAULT_INSTANCE.toBuilder().mergeFrom(prototype);
  }
  @java.lang.Override
  public Builder toBuilder() {
    return this == DEFAULT_INSTANCE
        ? new Builder() : new Builder().mergeFrom(this);
  }

  @java.lang.Override
  protected Builder newBuilderForType(
      com.google.protobuf.GeneratedMessageV3.BuilderParent parent) {
    Builder builder = new Builder(parent);
    return builder;
  }
  /**
   * Protobuf type {@code policy.attributes.GetAttributesByValueFqnsRequest}
   */
  public static final class Builder extends
      com.google.protobuf.GeneratedMessageV3.Builder<Builder> implements
      // @@protoc_insertion_point(builder_implements:policy.attributes.GetAttributesByValueFqnsRequest)
      io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequestOrBuilder {
    public static final com.google.protobuf.Descriptors.Descriptor
        getDescriptor() {
      return io.opentdf.platform.policy.attributes.AttributesProto.internal_static_policy_attributes_GetAttributesByValueFqnsRequest_descriptor;
    }

    @java.lang.Override
    protected com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
        internalGetFieldAccessorTable() {
      return io.opentdf.platform.policy.attributes.AttributesProto.internal_static_policy_attributes_GetAttributesByValueFqnsRequest_fieldAccessorTable
          .ensureFieldAccessorsInitialized(
              io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest.class, io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest.Builder.class);
    }

    // Construct using io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest.newBuilder()
    private Builder() {

    }

    private Builder(
        com.google.protobuf.GeneratedMessageV3.BuilderParent parent) {
      super(parent);

    }
    @java.lang.Override
    public Builder clear() {
      super.clear();
      bitField0_ = 0;
      fqns_ =
          com.google.protobuf.LazyStringArrayList.emptyList();
      return this;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.Descriptor
        getDescriptorForType() {
      return io.opentdf.platform.policy.attributes.AttributesProto.internal_static_policy_attributes_GetAttributesByValueFqnsRequest_descriptor;
    }

    @java.lang.Override
    public io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest getDefaultInstanceForType() {
      return io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest.getDefaultInstance();
    }

    @java.lang.Override
    public io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest build() {
      io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest result = buildPartial();
      if (!result.isInitialized()) {
        throw newUninitializedMessageException(result);
      }
      return result;
    }

    @java.lang.Override
    public io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest buildPartial() {
      io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest result = new io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest(this);
      if (bitField0_ != 0) { buildPartial0(result); }
      onBuilt();
      return result;
    }

    private void buildPartial0(io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest result) {
      int from_bitField0_ = bitField0_;
      if (((from_bitField0_ & 0x00000001) != 0)) {
        fqns_.makeImmutable();
        result.fqns_ = fqns_;
      }
    }

    @java.lang.Override
    public Builder clone() {
      return super.clone();
    }
    @java.lang.Override
    public Builder setField(
        com.google.protobuf.Descriptors.FieldDescriptor field,
        java.lang.Object value) {
      return super.setField(field, value);
    }
    @java.lang.Override
    public Builder clearField(
        com.google.protobuf.Descriptors.FieldDescriptor field) {
      return super.clearField(field);
    }
    @java.lang.Override
    public Builder clearOneof(
        com.google.protobuf.Descriptors.OneofDescriptor oneof) {
      return super.clearOneof(oneof);
    }
    @java.lang.Override
    public Builder setRepeatedField(
        com.google.protobuf.Descriptors.FieldDescriptor field,
        int index, java.lang.Object value) {
      return super.setRepeatedField(field, index, value);
    }
    @java.lang.Override
    public Builder addRepeatedField(
        com.google.protobuf.Descriptors.FieldDescriptor field,
        java.lang.Object value) {
      return super.addRepeatedField(field, value);
    }
    @java.lang.Override
    public Builder mergeFrom(com.google.protobuf.Message other) {
      if (other instanceof io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest) {
        return mergeFrom((io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest)other);
      } else {
        super.mergeFrom(other);
        return this;
      }
    }

    public Builder mergeFrom(io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest other) {
      if (other == io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest.getDefaultInstance()) return this;
      if (!other.fqns_.isEmpty()) {
        if (fqns_.isEmpty()) {
          fqns_ = other.fqns_;
          bitField0_ |= 0x00000001;
        } else {
          ensureFqnsIsMutable();
          fqns_.addAll(other.fqns_);
        }
        onChanged();
      }
      this.mergeUnknownFields(other.getUnknownFields());
      onChanged();
      return this;
    }

    @java.lang.Override
    public final boolean isInitialized() {
      return true;
    }

    @java.lang.Override
    public Builder mergeFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws java.io.IOException {
      if (extensionRegistry == null) {
        throw new java.lang.NullPointerException();
      }
      try {
        boolean done = false;
        while (!done) {
          int tag = input.readTag();
          switch (tag) {
            case 0:
              done = true;
              break;
            case 10: {
              java.lang.String s = input.readStringRequireUtf8();
              ensureFqnsIsMutable();
              fqns_.add(s);
              break;
            } // case 10
            default: {
              if (!super.parseUnknownField(input, extensionRegistry, tag)) {
                done = true; // was an endgroup tag
              }
              break;
            } // default:
          } // switch (tag)
        } // while (!done)
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        throw e.unwrapIOException();
      } finally {
        onChanged();
      } // finally
      return this;
    }
    private int bitField0_;

    private com.google.protobuf.LazyStringArrayList fqns_ =
        com.google.protobuf.LazyStringArrayList.emptyList();
    private void ensureFqnsIsMutable() {
      if (!fqns_.isModifiable()) {
        fqns_ = new com.google.protobuf.LazyStringArrayList(fqns_);
      }
      bitField0_ |= 0x00000001;
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @return A list containing the fqns.
     */
    public com.google.protobuf.ProtocolStringList
        getFqnsList() {
      fqns_.makeImmutable();
      return fqns_;
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @return The count of fqns.
     */
    public int getFqnsCount() {
      return fqns_.size();
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @param index The index of the element to return.
     * @return The fqns at the given index.
     */
    public java.lang.String getFqns(int index) {
      return fqns_.get(index);
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @param index The index of the value to return.
     * @return The bytes of the fqns at the given index.
     */
    public com.google.protobuf.ByteString
        getFqnsBytes(int index) {
      return fqns_.getByteString(index);
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @param index The index to set the value at.
     * @param value The fqns to set.
     * @return This builder for chaining.
     */
    public Builder setFqns(
        int index, java.lang.String value) {
      if (value == null) { throw new NullPointerException(); }
      ensureFqnsIsMutable();
      fqns_.set(index, value);
      bitField0_ |= 0x00000001;
      onChanged();
      return this;
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @param value The fqns to add.
     * @return This builder for chaining.
     */
    public Builder addFqns(
        java.lang.String value) {
      if (value == null) { throw new NullPointerException(); }
      ensureFqnsIsMutable();
      fqns_.add(value);
      bitField0_ |= 0x00000001;
      onChanged();
      return this;
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @param values The fqns to add.
     * @return This builder for chaining.
     */
    public Builder addAllFqns(
        java.lang.Iterable<java.lang.String> values) {
      ensureFqnsIsMutable();
      com.google.protobuf.AbstractMessageLite.Builder.addAll(
          values, fqns_);
      bitField0_ |= 0x00000001;
      onChanged();
      return this;
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @return This builder for chaining.
     */
    public Builder clearFqns() {
      fqns_ =
        com.google.protobuf.LazyStringArrayList.emptyList();
      bitField0_ = (bitField0_ & ~0x00000001);;
      onChanged();
      return this;
    }
    /**
     * <code>repeated string fqns = 1 [json_name = "fqns", (.buf.validate.field) = { ... }</code>
     * @param value The bytes of the fqns to add.
     * @return This builder for chaining.
     */
    public Builder addFqnsBytes(
        com.google.protobuf.ByteString value) {
      if (value == null) { throw new NullPointerException(); }
      checkByteStringIsUtf8(value);
      ensureFqnsIsMutable();
      fqns_.add(value);
      bitField0_ |= 0x00000001;
      onChanged();
      return this;
    }
    @java.lang.Override
    public final Builder setUnknownFields(
        final com.google.protobuf.UnknownFieldSet unknownFields) {
      return super.setUnknownFields(unknownFields);
    }

    @java.lang.Override
    public final Builder mergeUnknownFields(
        final com.google.protobuf.UnknownFieldSet unknownFields) {
      return super.mergeUnknownFields(unknownFields);
    }


    // @@protoc_insertion_point(builder_scope:policy.attributes.GetAttributesByValueFqnsRequest)
  }

  // @@protoc_insertion_point(class_scope:policy.attributes.GetAttributesByValueFqnsRequest)
  private static final io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest DEFAULT_INSTANCE;
  static {
    DEFAULT_INSTANCE = new io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest();
  }

  public static io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest getDefaultInstance() {
    return DEFAULT_INSTANCE;
  }

  private static final com.google.protobuf.Parser<GetAttributesByValueFqnsRequest>
      PARSER = new com.google.protobuf.AbstractParser<GetAttributesByValueFqnsRequest>() {
    @java.lang.Override
    public GetAttributesByValueFqnsRequest parsePartialFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws com.google.protobuf.InvalidProtocolBufferException {
      Builder builder = newBuilder();
      try {
        builder.mergeFrom(input, extensionRegistry);
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        throw e.setUnfinishedMessage(builder.buildPartial());
      } catch (com.google.protobuf.UninitializedMessageException e) {
        throw e.asInvalidProtocolBufferException().setUnfinishedMessage(builder.buildPartial());
      } catch (java.io.IOException e) {
        throw new com.google.protobuf.InvalidProtocolBufferException(e)
            .setUnfinishedMessage(builder.buildPartial());
      }
      return builder.buildPartial();
    }
  };

  public static com.google.protobuf.Parser<GetAttributesByValueFqnsRequest> parser() {
    return PARSER;
  }

  @java.lang.Override
  public com.google.protobuf.Parser<GetAttributesByValueFqnsRequest> getParserForType() {
    return PARSER;
  }

  @java.lang.Override
  public io.opentdf.platform.policy.attributes.GetAttributesByValueFqnsRequest getDefaultInstanceForType() {
    return DEFAULT_INSTANCE;
  }

}

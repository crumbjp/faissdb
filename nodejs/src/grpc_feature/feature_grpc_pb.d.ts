// package: feature
// file: feature.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as feature_pb from "./feature_pb";

interface IFeatureService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    status: IFeatureService_IStatus;
    set: IFeatureService_ISet;
    del: IFeatureService_IDel;
    search: IFeatureService_ISearch;
    train: IFeatureService_ITrain;
    dropall: IFeatureService_IDropall;
    dbStats: IFeatureService_IDbStats;
}

interface IFeatureService_IStatus extends grpc.MethodDefinition<feature_pb.StatusRequest, feature_pb.StatusReply> {
    path: "/feature.Feature/Status";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<feature_pb.StatusRequest>;
    requestDeserialize: grpc.deserialize<feature_pb.StatusRequest>;
    responseSerialize: grpc.serialize<feature_pb.StatusReply>;
    responseDeserialize: grpc.deserialize<feature_pb.StatusReply>;
}
interface IFeatureService_ISet extends grpc.MethodDefinition<feature_pb.SetRequest, feature_pb.SetReply> {
    path: "/feature.Feature/Set";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<feature_pb.SetRequest>;
    requestDeserialize: grpc.deserialize<feature_pb.SetRequest>;
    responseSerialize: grpc.serialize<feature_pb.SetReply>;
    responseDeserialize: grpc.deserialize<feature_pb.SetReply>;
}
interface IFeatureService_IDel extends grpc.MethodDefinition<feature_pb.DelRequest, feature_pb.DelReply> {
    path: "/feature.Feature/Del";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<feature_pb.DelRequest>;
    requestDeserialize: grpc.deserialize<feature_pb.DelRequest>;
    responseSerialize: grpc.serialize<feature_pb.DelReply>;
    responseDeserialize: grpc.deserialize<feature_pb.DelReply>;
}
interface IFeatureService_ISearch extends grpc.MethodDefinition<feature_pb.SearchRequest, feature_pb.SearchReply> {
    path: "/feature.Feature/Search";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<feature_pb.SearchRequest>;
    requestDeserialize: grpc.deserialize<feature_pb.SearchRequest>;
    responseSerialize: grpc.serialize<feature_pb.SearchReply>;
    responseDeserialize: grpc.deserialize<feature_pb.SearchReply>;
}
interface IFeatureService_ITrain extends grpc.MethodDefinition<feature_pb.TrainRequest, feature_pb.TrainReply> {
    path: "/feature.Feature/Train";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<feature_pb.TrainRequest>;
    requestDeserialize: grpc.deserialize<feature_pb.TrainRequest>;
    responseSerialize: grpc.serialize<feature_pb.TrainReply>;
    responseDeserialize: grpc.deserialize<feature_pb.TrainReply>;
}
interface IFeatureService_IDropall extends grpc.MethodDefinition<feature_pb.DropallRequest, feature_pb.DropallReply> {
    path: "/feature.Feature/Dropall";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<feature_pb.DropallRequest>;
    requestDeserialize: grpc.deserialize<feature_pb.DropallRequest>;
    responseSerialize: grpc.serialize<feature_pb.DropallReply>;
    responseDeserialize: grpc.deserialize<feature_pb.DropallReply>;
}
interface IFeatureService_IDbStats extends grpc.MethodDefinition<feature_pb.DbStatsRequest, feature_pb.DbStatsReply> {
    path: "/feature.Feature/DbStats";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<feature_pb.DbStatsRequest>;
    requestDeserialize: grpc.deserialize<feature_pb.DbStatsRequest>;
    responseSerialize: grpc.serialize<feature_pb.DbStatsReply>;
    responseDeserialize: grpc.deserialize<feature_pb.DbStatsReply>;
}

export const FeatureService: IFeatureService;

export interface IFeatureServer extends grpc.UntypedServiceImplementation {
    status: grpc.handleUnaryCall<feature_pb.StatusRequest, feature_pb.StatusReply>;
    set: grpc.handleUnaryCall<feature_pb.SetRequest, feature_pb.SetReply>;
    del: grpc.handleUnaryCall<feature_pb.DelRequest, feature_pb.DelReply>;
    search: grpc.handleUnaryCall<feature_pb.SearchRequest, feature_pb.SearchReply>;
    train: grpc.handleUnaryCall<feature_pb.TrainRequest, feature_pb.TrainReply>;
    dropall: grpc.handleUnaryCall<feature_pb.DropallRequest, feature_pb.DropallReply>;
    dbStats: grpc.handleUnaryCall<feature_pb.DbStatsRequest, feature_pb.DbStatsReply>;
}

export interface IFeatureClient {
    status(request: feature_pb.StatusRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.StatusReply) => void): grpc.ClientUnaryCall;
    status(request: feature_pb.StatusRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.StatusReply) => void): grpc.ClientUnaryCall;
    status(request: feature_pb.StatusRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.StatusReply) => void): grpc.ClientUnaryCall;
    set(request: feature_pb.SetRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.SetReply) => void): grpc.ClientUnaryCall;
    set(request: feature_pb.SetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.SetReply) => void): grpc.ClientUnaryCall;
    set(request: feature_pb.SetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.SetReply) => void): grpc.ClientUnaryCall;
    del(request: feature_pb.DelRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.DelReply) => void): grpc.ClientUnaryCall;
    del(request: feature_pb.DelRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.DelReply) => void): grpc.ClientUnaryCall;
    del(request: feature_pb.DelRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.DelReply) => void): grpc.ClientUnaryCall;
    search(request: feature_pb.SearchRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.SearchReply) => void): grpc.ClientUnaryCall;
    search(request: feature_pb.SearchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.SearchReply) => void): grpc.ClientUnaryCall;
    search(request: feature_pb.SearchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.SearchReply) => void): grpc.ClientUnaryCall;
    train(request: feature_pb.TrainRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.TrainReply) => void): grpc.ClientUnaryCall;
    train(request: feature_pb.TrainRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.TrainReply) => void): grpc.ClientUnaryCall;
    train(request: feature_pb.TrainRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.TrainReply) => void): grpc.ClientUnaryCall;
    dropall(request: feature_pb.DropallRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.DropallReply) => void): grpc.ClientUnaryCall;
    dropall(request: feature_pb.DropallRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.DropallReply) => void): grpc.ClientUnaryCall;
    dropall(request: feature_pb.DropallRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.DropallReply) => void): grpc.ClientUnaryCall;
    dbStats(request: feature_pb.DbStatsRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.DbStatsReply) => void): grpc.ClientUnaryCall;
    dbStats(request: feature_pb.DbStatsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.DbStatsReply) => void): grpc.ClientUnaryCall;
    dbStats(request: feature_pb.DbStatsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.DbStatsReply) => void): grpc.ClientUnaryCall;
}

export class FeatureClient extends grpc.Client implements IFeatureClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public status(request: feature_pb.StatusRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.StatusReply) => void): grpc.ClientUnaryCall;
    public status(request: feature_pb.StatusRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.StatusReply) => void): grpc.ClientUnaryCall;
    public status(request: feature_pb.StatusRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.StatusReply) => void): grpc.ClientUnaryCall;
    public set(request: feature_pb.SetRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.SetReply) => void): grpc.ClientUnaryCall;
    public set(request: feature_pb.SetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.SetReply) => void): grpc.ClientUnaryCall;
    public set(request: feature_pb.SetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.SetReply) => void): grpc.ClientUnaryCall;
    public del(request: feature_pb.DelRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.DelReply) => void): grpc.ClientUnaryCall;
    public del(request: feature_pb.DelRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.DelReply) => void): grpc.ClientUnaryCall;
    public del(request: feature_pb.DelRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.DelReply) => void): grpc.ClientUnaryCall;
    public search(request: feature_pb.SearchRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.SearchReply) => void): grpc.ClientUnaryCall;
    public search(request: feature_pb.SearchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.SearchReply) => void): grpc.ClientUnaryCall;
    public search(request: feature_pb.SearchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.SearchReply) => void): grpc.ClientUnaryCall;
    public train(request: feature_pb.TrainRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.TrainReply) => void): grpc.ClientUnaryCall;
    public train(request: feature_pb.TrainRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.TrainReply) => void): grpc.ClientUnaryCall;
    public train(request: feature_pb.TrainRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.TrainReply) => void): grpc.ClientUnaryCall;
    public dropall(request: feature_pb.DropallRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.DropallReply) => void): grpc.ClientUnaryCall;
    public dropall(request: feature_pb.DropallRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.DropallReply) => void): grpc.ClientUnaryCall;
    public dropall(request: feature_pb.DropallRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.DropallReply) => void): grpc.ClientUnaryCall;
    public dbStats(request: feature_pb.DbStatsRequest, callback: (error: grpc.ServiceError | null, response: feature_pb.DbStatsReply) => void): grpc.ClientUnaryCall;
    public dbStats(request: feature_pb.DbStatsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: feature_pb.DbStatsReply) => void): grpc.ClientUnaryCall;
    public dbStats(request: feature_pb.DbStatsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: feature_pb.DbStatsReply) => void): grpc.ClientUnaryCall;
}

export namespace model {
	
	export class GitCommit {
	    repoName: string;
	    repoPath: string;
	    hash: string;
	    authorName: string;
	    authorEmail: string;
	    commitTime: string;
	    message: string;
	    files: string[];
	
	    static createFrom(source: any = {}) {
	        return new GitCommit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repoName = source["repoName"];
	        this.repoPath = source["repoPath"];
	        this.hash = source["hash"];
	        this.authorName = source["authorName"];
	        this.authorEmail = source["authorEmail"];
	        this.commitTime = source["commitTime"];
	        this.message = source["message"];
	        this.files = source["files"];
	    }
	}
	export class RepoInfo {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class GitLogRequest {
	    repos: RepoInfo[];
	    since: string;
	    until: string;
	    author: string;
	    noMerges: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GitLogRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repos = this.convertValues(source["repos"], RepoInfo);
	        this.since = source["since"];
	        this.until = source["until"];
	        this.author = source["author"];
	        this.noMerges = source["noMerges"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RepoError {
	    repoName: string;
	    repoPath: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repoName = source["repoName"];
	        this.repoPath = source["repoPath"];
	        this.message = source["message"];
	    }
	}
	export class GitLogResponse {
	    commits: GitCommit[];
	    errors: RepoError[];
	
	    static createFrom(source: any = {}) {
	        return new GitLogResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.commits = this.convertValues(source["commits"], GitCommit);
	        this.errors = this.convertValues(source["errors"], RepoError);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class ScanRequest {
	    rootPath: string;
	    maxDepth: number;
	    excludeRepoNames: string[];
	
	    static createFrom(source: any = {}) {
	        return new ScanRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootPath = source["rootPath"];
	        this.maxDepth = source["maxDepth"];
	        this.excludeRepoNames = source["excludeRepoNames"];
	    }
	}

}


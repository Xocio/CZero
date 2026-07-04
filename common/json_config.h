#ifndef CZERO_JSON_CONFIG_H
#define CZERO_JSON_CONFIG_H
#include <string>
#include <unordered_map>
#include <fstream>
#include <iterator>
#include <cstddef>

namespace czero {

inline void jsonSkipWs(const std::string& s, size_t& i) {
    while (i < s.size() && (s[i]==' '||s[i]=='\t'||s[i]=='\n'||s[i]=='\r')) i++;
}

inline std::string jsonParseString(const std::string& s, size_t& i) {
    std::string out; i++;
    while (i < s.size() && s[i] != '"') {
        if (s[i]=='\\' && i+1 < s.size()) { i++; char c=s[i];
            if (c=='n') out+='\n'; else if (c=='t') out+='\t'; else if (c=='r') out+='\r'; else out+=c; }
        else out += s[i];
        i++;
    }
    if (i < s.size()) i++;
    return out;
}

inline void jsonParseObject(const std::string& s, size_t& i, const std::string& prefix,
                            std::unordered_map<std::string,std::string>& out) {
    i++; jsonSkipWs(s, i);
    if (i < s.size() && s[i]=='}') { i++; return; }
    while (i < s.size()) {
        jsonSkipWs(s, i);
        if (i >= s.size() || s[i] != '"') break;
        std::string key = jsonParseString(s, i);
        std::string full = prefix.empty() ? key : prefix + "." + key;
        jsonSkipWs(s, i);
        if (i < s.size() && s[i]==':') i++;
        jsonSkipWs(s, i);
        if (i >= s.size()) break;
        if (s[i]=='{') jsonParseObject(s, i, full, out);
        else if (s[i]=='"') out[full] = jsonParseString(s, i);
        else if (s[i]=='[') { int d=0; do { if(s[i]=='[')d++; else if(s[i]==']')d--; i++; } while(i<s.size()&&d>0); }
        else { std::string v; while (i<s.size() && s[i]!=','&&s[i]!='}'&&s[i]!=' '&&s[i]!='\t'&&s[i]!='\n'&&s[i]!='\r') { v+=s[i]; i++; } out[full]=v; }
        jsonSkipWs(s, i);
        if (i < s.size() && s[i]==',') { i++; continue; }
        if (i < s.size() && s[i]=='}') { i++; break; }
    }
}

inline std::unordered_map<std::string,std::string> parseJsonConfig(const std::string& path) {
    std::unordered_map<std::string,std::string> out;
    std::ifstream f(path); if (!f.is_open()) return out;
    std::string s((std::istreambuf_iterator<char>(f)), std::istreambuf_iterator<char>());
    size_t i = 0; jsonSkipWs(s, i);
    if (i < s.size() && s[i]=='{') jsonParseObject(s, i, "", out);
    return out;
}

} 

#endif

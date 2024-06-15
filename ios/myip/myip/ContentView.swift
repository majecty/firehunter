//
//  ContentView.swift
//  myip
//
//  Created by 주형 on 6/15/24.
//

import SwiftUI

struct ContentView: View {
    @State var myAddress: String?
    @State var myAddresses = [String]()
    
    var body: some View {
        VStack {
            Image(systemName: "globe")
                .imageScale(.large)
                .foregroundStyle(.tint)
            Text("Hello, world!")
            Button(action: getMyIp) {
                Text("내 정보 조회")
            }
            Text(myAddress ?? "No IP")
        }
        .padding()
    }
    
    func getMyIp() {
        self.myAddress = nil
        self.myAddresses = []
        
        var address: String?
        
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        guard getifaddrs(&ifaddr) == 0 else {
            myAddress = "Cannot Get MyIp(1)"
            return
        }
        guard let firstAddr = ifaddr else {
            myAddress = "Cannot Get MyIp(2)"
            return
        }
        for ifptr in sequence(first: firstAddr, next: { $0.pointee.ifa_next}) {
            let interface = ifptr.pointee
            
            let addrFamily = interface.ifa_addr.pointee.sa_family
            if addrFamily == UInt8(AF_INET) || addrFamily == UInt8(AF_INET6) {
                let name = String(cString: interface.ifa_name)
                if name == "en0" {
                    var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                    getnameinfo(interface.ifa_addr, socklen_t(interface.ifa_addr.pointee.sa_len), &hostname, socklen_t(hostname.count), nil, socklen_t(0), NI_NUMERICHOST)
                    address = String(cString: hostname)
                } else {
                    var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                    getnameinfo(interface.ifa_addr, socklen_t(interface.ifa_addr.pointee.sa_len), &hostname, socklen_t(hostname.count), nil, socklen_t(0), NI_NUMERICHOST)
                    self.myAddresses.append(String(cString: hostname))
                }
            }
        }
        
        freeifaddrs(ifaddr)
        self.myAddress = address
    }
}

#Preview {
    ContentView()
}
